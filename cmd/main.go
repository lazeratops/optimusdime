package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/lazeratops/optimusdime/src/converter/exchangeapi"
	"github.com/lazeratops/optimusdime/src/document"
	"github.com/lazeratops/optimusdime/src/importer"
	"github.com/lazeratops/optimusdime/src/llm"
	"github.com/lazeratops/optimusdime/src/parser"
)

func main() {
	csvPath := flag.String("statement", "", "Path to CSV file of bank statement")
	openaiApiKey := flag.String("oai_key", "", "OpenAI API Key")

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
	fmt.Println("OLD DOC")

	fmt.Println(doc)
	converterApi, err := exchangeapi.NewExchangeApi("")
	if err != nil {
		log.Fatal(err)
	}
	convertedDoc, failedDoc, err := converterApi.Convert(document.SEK, doc)
	if err != nil {
		log.Fatal(err)
	}
	fileName := filepath.Base(*csvPath)
	fmt.Println("CONVERTED DOC")
	fmt.Println(convertedDoc)
	err = convertedDoc.SaveToCSV(fmt.Sprintf("convered_%s", fileName))
	if err != nil {
		log.Fatal(err)
	}
	err = failedDoc.SaveToCSV(fmt.Sprintf("failed_%s", fileName))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("RESULTS:")
	log.Printf("\nTotal original transactions: %d", len(doc.Transactions))

	lSuccess := len(convertedDoc.Transactions)
	lFail := len(failedDoc.Transactions)
	log.Printf("\nTotal processed transactions: %d", lSuccess+lFail)

	log.Printf("\nSuccessfully converted: %d", lSuccess)
	log.Printf("\nFailed conversion: %d", lFail)

}
