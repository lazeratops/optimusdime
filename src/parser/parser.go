package parser

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/lazeratops/optimusdime/src/document"
	"github.com/lazeratops/optimusdime/src/llm"
	"github.com/lazeratops/optimusdime/src/util"
)

var ErrLLMFail = errors.New("LLM call failed")

type Parser struct {
	llm llm.Llm
}

func NewParser(llm llm.Llm) *Parser {
	return &Parser{
		llm: llm,
	}
}

func (p *Parser) Parse(records [][]string) (*document.Document, error) {
	var content strings.Builder
	for _, record := range records {
		content.WriteString(strings.Join(record, ",") + "\n")
	}

	indices, err := p.llm.FindElements(llm.DesiredElements{
		"date":        "The date of the transaction",
		"amount":      "The monetary amount of the transaction",
		"currency":    "The currency the transaction was performed in",
		"description": "The description of the transaction",
	}, content.String())
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, ErrLLMFail)
	}
	var transactions []document.Transaction
	for _, record := range records {
		date, err := util.ParseDate(record[indices["date"]])
		if err != nil {
			log.Printf("\nfailed to parse date: %v; skipping", err)
			continue
		}

		amount, err := strconv.ParseFloat(record[indices["amount"]], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %w", err)
		}

		c := record[indices["currency"]]
		currency := document.Currency(c)

		description := record[indices["description"]]

		transactions = append(transactions, document.Transaction{
			Date:        date,
			Amount:      amount,
			Currency:    currency,
			Description: description,
		})
	}

	return &document.Document{
		Transactions: transactions,
	}, nil
}
