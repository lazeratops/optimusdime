package importer

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/lazeratops/optimusdime/src/document"
	"github.com/lazeratops/optimusdime/src/parser"
)

type Csv struct {
	parser *parser.Parser
}

type CsvConfig struct {
	delimiter rune
}

func NewCsv(parser *parser.Parser) *Csv {
	return &Csv{
		parser: parser,
	}
}

func (c *Csv) Import(filePath string, config *CsvConfig) (*document.Document, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV: %w", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	if config != nil {
		reader.Comma = config.delimiter
	}
	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("csv file is empty")
	}
	return c.parser.Parse(records)
}
