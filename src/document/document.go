package document

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Currency string

const (
	USD Currency = "USD"
	SEK Currency = "SEK"
	EUR Currency = "EUR"
)

type Document struct {
	Transactions []Transaction `json:"transactions" jsonschema_description:"All bank transactions in the document"`
}

type Transaction struct {
	Description string    `json:"description" jsonschema_description:"The description of the transaction"`
	Date        time.Time `json:"date" jsonschema_description:"The date of the transaction"`
	Amount      float64   `json:"amount" jsonschema_description:"The amount of the transaction"`
	Currency    Currency  `json:"currency" jsonschema_description:"The currency of the transaction"`
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction
	aux := struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse the date
	parsedDate, err := time.Parse("02-01-2006", aux.Date)
	if err != nil {
		return fmt.Errorf("failed to parse date %s: %w", aux.Date, err)
	}

	t.Date = parsedDate
	return nil
}

func (d *Document) SaveToCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Date", "Description", "Amount", "Currency"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, t := range d.Transactions {
		record := []string{
			t.Date.Format("2006-01-02"),
			t.Description,
			fmt.Sprintf("%.2f", t.Amount),
			string(t.Currency),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}
