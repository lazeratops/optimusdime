package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/lazeratops/optimusdime/src/converter/exchangeapi"
	"github.com/lazeratops/optimusdime/src/document"
	"github.com/lazeratops/optimusdime/src/importer"
	"github.com/lazeratops/optimusdime/src/llm"
	"github.com/lazeratops/optimusdime/src/parser"
)

const resultsBanner = `
╔═══════════════════════════════════════════════════════╗
║                  CONVERSION RESULTS                   ║
╚═══════════════════════════════════════════════════════╝
`

func main() {
	csvPath := flag.String("statement", "", "Path to CSV file of bank statement")
	openaiApiKey := flag.String("oai_key", "", "OpenAI API Key")
	targetCurrency := flag.String("target_currenct", "SEK", "Target currency")

	flag.Parse()

	if *csvPath == "" {
		log.Fatal("Please provide a file path using -statement flag")
	}
	if *openaiApiKey == "" {
		log.Fatal("Please provide an OpenAI API key using the -oai_key flag")
	}

	llm, err := llm.NewOpenAi(llm.Config{
		ApiKey: *openaiApiKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	parser := parser.NewParser(llm)
	importer := importer.NewCsv(parser)

	doc, err := importer.Import(*csvPath, nil)
	if err != nil {
		log.Fatal(err)
	}
	converterApi, err := exchangeapi.NewExchangeApi("")
	if err != nil {
		log.Fatal(err)
	}
	convertedDoc, failedDoc, err := converterApi.Convert(document.Currency(*targetCurrency), doc)
	if err != nil {
		log.Fatal(err)
	}
	fileName := filepath.Base(*csvPath)

	successFilename := fmt.Sprintf("convered_%s", fileName)
	failedFilename := fmt.Sprintf("failed_%s", fileName)
	err = convertedDoc.SaveToCSV(successFilename)
	if err != nil {
		log.Fatal(err)
	}
	err = failedDoc.SaveToCSV(failedFilename)
	if err != nil {
		log.Fatal(err)
	}
	println(resultsBanner)
	println(fmt.Sprintf("Target Currency: %s", *targetCurrency))
	println(fmt.Sprintf("- %s", successFilename))
	println(fmt.Sprintf("- %s", failedFilename))
	println()
	lSuccess := len(convertedDoc.Transactions)
	lFail := len(failedDoc.Transactions)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Total Transactions", "Total Processed", "Succeeded #", "Failed #"})
	t.AppendRows([]table.Row{
		{len(doc.Transactions), lSuccess + lFail, lSuccess, lFail},
	})
	t.AppendSeparator()
	t.SetStyle(table.StyleBold)
	t.Render()
}
