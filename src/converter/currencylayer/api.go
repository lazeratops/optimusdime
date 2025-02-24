package currencylayer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lazeratops/optimusdime/src/converter"
	"github.com/lazeratops/optimusdime/src/document"
)

const defaultApiUrl = "api.currencylayer.com/historical"

type Api struct {
	apiKey string
	url    string
	schema string
}

func NewCurrencyLayer(apiUrl string, apiKey string) (*Api, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key must be provided for CurrencyLayer API")
	}
	var schema string
	if apiUrl == "" {
		apiUrl = defaultApiUrl
		schema = "https"
	} else {

		// Parse URL
		parsedURL, err := url.Parse(apiUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid API URL: %w", err)
		}
		if parsedURL.Scheme != "" {
			schema = parsedURL.Scheme
		}
		apiUrl = parsedURL.Host
	}

	return &Api{
		url:    apiUrl,
		schema: schema,
		apiKey: apiKey,
	}, nil
}

func (api *Api) Convert(targetCurrency document.Currency, statement *document.Document) (*document.Document, *document.Document, error) {
	if len(statement.Transactions) == 0 {
		return nil, nil, errors.New("no transactions to convert")
	}

	newDoc := document.Document{
		Transactions: []document.Transaction{},
	}

	failedToConvertDoc := document.Document{
		Transactions: []document.Transaction{},
	}

	type dateCurrenciesTransactions struct {
		currencies   []document.Currency
		transactions []document.Transaction
	}

	sourceCurrenciesForDate := make(map[time.Time]dateCurrenciesTransactions)

	for _, oldTransaction := range statement.Transactions {
		sourceCurrenciesForDate[oldTransaction.Date] = dateCurrenciesTransactions{
			transactions: append(sourceCurrenciesForDate[oldTransaction.Date].transactions, oldTransaction),
			currencies:   append(sourceCurrenciesForDate[oldTransaction.Date].currencies, oldTransaction.Currency),
		}
	}

	var lastError error
	for date, v := range sourceCurrenciesForDate {
		c := document.Currency(targetCurrency)
		url := api.getUrl()
		if url == "" {
			return nil, nil, fmt.Errorf("failed to get api URL for date %v and currency %v", date, c)
		}
		v.currencies = append(v.currencies, targetCurrency)
		resBody, err := api.fetch(url, v.currencies, date)
		if err != nil {
			failedToConvertDoc.Transactions = append(failedToConvertDoc.Transactions, v.transactions...)
			log.Printf("\n Failed to fetch transactions from URL: %s: %v", url, err)
			lastError = err
			continue
		}

		var currencyRes ApiResponse
		if err := json.Unmarshal(resBody, &currencyRes); err != nil {
			failedToConvertDoc.Transactions = append(failedToConvertDoc.Transactions, v.transactions...)
			log.Printf("\n Failed to parse currency response from URL: %s: %v", url, err)
			lastError = err
			continue
		}

		if currencyRes.Source != document.USD {
			return nil, nil, fmt.Errorf("unexpected currency response from CurrencyLayer. Expected USD source, got %s", currencyRes.Source)
		}

		for _, oldTransaction := range v.transactions {
			rate, err := currencyRes.getCrossRate(oldTransaction.Currency, targetCurrency)
			if err != nil {
				failedToConvertDoc.Transactions = append(failedToConvertDoc.Transactions, oldTransaction)
				log.Printf("\nFailed to get cross rate for %s to %s: %v", oldTransaction.Currency, targetCurrency, err)
				continue
			}

			convertedAmount := oldTransaction.Amount * rate
			convertedAmount = math.Round(convertedAmount*100) / 100

			transaction := &document.Transaction{
				Description: oldTransaction.Description,
				Currency:    targetCurrency,
				Date:        oldTransaction.Date,
				Amount:      convertedAmount,
			}
			newDoc.Transactions = append(newDoc.Transactions, *transaction)
		}
	}
	if len(newDoc.Transactions) == 0 && lastError != nil {
		return nil, &failedToConvertDoc, lastError
	}
	return &newDoc, &failedToConvertDoc, nil
}

func (api *Api) getUrl() string {
	return fmt.Sprintf("%s://%s", api.schema, api.url)
}

func (api *Api) fetch(apiUrl string, currencies []document.Currency, date time.Time) ([]byte, error) {
	// Create base URL
	baseURL, err := url.Parse(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CurrencyLayer API URL: %w", err)
	}

	params := baseURL.Query()
	params.Add("access_key", api.apiKey)
	params.Add("date", date.Format("2006-01-02"))

	if len(currencies) > 0 {
		currencyStrs := make([]string, len(currencies))
		for i, c := range currencies {
			currencyStrs[i] = string(c)
		}
		params.Add("currencies", strings.Join(currencyStrs, ","))
	}

	baseURL.RawQuery = params.Encode()

	// Make the request
	resp, err := http.Get(baseURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned error %d: %s: %w", resp.StatusCode, resp.Status, converter.ErrFailedAPICall)
	}

	return body, nil
}
