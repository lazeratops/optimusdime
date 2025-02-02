package exchangeapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lazeratops/optimusdime/src/document"
)

var ErrFailedAPICall = errors.New("bad response from currency exchange API")

const (
	//exchangeApiUrl         = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1/"
	defaultApiUrl = "currency-api.pages.dev/v1/currencies"
)

type Api struct {
	url    string
	schema string
}

func NewExchangeApi(apiUrl string) (*Api, error) {
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
	}, nil
}

func (api *Api) getUrl(date time.Time, targetCurrency document.Currency) string {
	dateStr := date.Format("2006-01-02")
	c := string(targetCurrency)
	if c == "" {
		return ""
	}
	lower := strings.ToLower(c)
	if api.url != defaultApiUrl {
		return fmt.Sprintf("%s://%s/%s.json", api.schema, api.url, lower)
	}
	return fmt.Sprintf("%s://%s.%s/%s.json", api.schema, dateStr, api.url, lower)
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

	sourceCurrenciesForDate := make(map[time.Time][]document.Transaction)

	for _, oldTransaction := range statement.Transactions {
		sourceCurrenciesForDate[oldTransaction.Date] = append(sourceCurrenciesForDate[oldTransaction.Date], oldTransaction)
	}

	for date, transactions := range sourceCurrenciesForDate {
		c := document.Currency(targetCurrency)
		url := api.getUrl(date, c)
		if url == "" {
			return nil, nil, fmt.Errorf("failed to get api URL for date %v and currency %v", date, c)
		}
		resBody, err := api.fetch(url)
		if err != nil {
			failedToConvertDoc.Transactions = append(failedToConvertDoc.Transactions, transactions...)
			log.Printf("\n Failed to fetch transactions from URL: %s: %v", url, err)
			continue
		}

		var currencyRes ApiResponse
		if err := json.Unmarshal(resBody, &currencyRes); err != nil {
			failedToConvertDoc.Transactions = append(failedToConvertDoc.Transactions, transactions...)
			log.Printf("\n Failed to parse currency response from URL: %s: %v", url, err)
			continue
		}

		sTargetCurrency := strings.ToLower(string(targetCurrency))
		rates, ok := currencyRes.Rates[strings.ToLower(sTargetCurrency)]
		if !ok {
			failedToConvertDoc.Transactions = append(failedToConvertDoc.Transactions, transactions...)
			log.Printf("\n Failed to get currency rates for target currenct from URL: %s: %v", url, err)
			continue
		}

		for _, oldTransaction := range transactions {
			sSourceCurrency := strings.ToLower(string(oldTransaction.Currency))
			rate, ok := rates[sSourceCurrency]
			if !ok {
				failedToConvertDoc.Transactions = append(failedToConvertDoc.Transactions, oldTransaction)
				log.Printf("\n Failed to get currency rates for transaction %s: %v", oldTransaction.Description, err)
				continue
			}
			convertedAmount := oldTransaction.Amount / rate

			transaction := &document.Transaction{
				Description: oldTransaction.Description,
				Currency:    targetCurrency,
				Date:        oldTransaction.Date,
				Amount:      convertedAmount,
			}
			newDoc.Transactions = append(newDoc.Transactions, *transaction)
		}
	}

	return &newDoc, &failedToConvertDoc, nil
}

func (api *Api) fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned error %d: %s: %w", resp.StatusCode, resp.Status, ErrFailedAPICall)
	}

	return body, nil
}
