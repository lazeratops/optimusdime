package parsertest

import (
	"encoding/csv"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lazeratops/optimusdime/mocks"
	"github.com/lazeratops/optimusdime/src/document"
	"github.com/lazeratops/optimusdime/src/parser"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	date_30122024, err := time.Parse("02-01-2006", "30-12-2024")
	require.NoError(t, err)
	t.Parallel()
	cases := []struct {
		name    string
		doc     string
		llmRes  func(t *testing.T) (map[string]int, error)
		wantDoc document.Document
		wantErr error
	}{
		{name: "error-llmres",
			doc: `"TransferWise ID",Date,Amount,Currency,Description,"Payment Reference","Running Balance","Exchange From","Exchange To","Exchange Rate","Payer Name","Payee Name","Payee Account Number",Merchant,"Card Last Four Digits","Card Holder Full Name",Attachment,Note,"Total fees","Exchange To Amount"
TRANSFER-1356004938,30-12-2024,17.76,USD,"Received money from AMAZON AUSTRALIA SERVICES  INC. with reference PAYMENT",PAYMENT,272.63,,,,"AMAZON AUSTRALIA SERVICES  INC.",,,,,,,,0.00,
TRANSFER-1356004879,30-12-2024,19.87,USD,"Received money from AMAZON.COM SERVICES LLC with reference PAYMENT",PAYMENT,254.87,,,,"AMAZON.COM SERVICES LLC",,,,,,,,0.00,
TRANSFER-1356003098,30-12-2024,12.33,USD,"Received money from AMAZON MEDIA EU S.A.R.L. with reference PAYMENT",PAYMENT,235.00,,,,"AMAZON MEDIA EU S.A.R.L.",,,,,,,,0.00,
CARD-2106097800,30-12-2024,-10.00,USD,"Card transaction of USD issued by Booksirens PHILADELPHIA",,222.67,,,,,,,"Booksirens PHILADELPHIA",1033,"Yelizaveta Shulyayeva",,,0.00,
TRANSFER-1355191945,30-12-2024,0.02,USD,"Received money from AMAZON SE5097806 with reference EDI PYMNTS","EDI PYMNTS",232.67,,,,"AMAZON SE5097806",,,,,,,,0.00,
`,
			llmRes: func(t *testing.T) (map[string]int, error) {
				return nil, errors.New("some error")
			},
			wantErr: parser.ErrLLMFail,
		},
		{
			name: "success",
			doc: `"TransferWise ID",Date,Amount,Currency,Description,"Payment Reference","Running Balance","Exchange From","Exchange To","Exchange Rate","Payer Name","Payee Name","Payee Account Number",Merchant,"Card Last Four Digits","Card Holder Full Name",Attachment,Note,"Total fees","Exchange To Amount"
TRANSFER-1356004938,30-12-2024,17.76,USD,"Received money from AMAZON AUSTRALIA SERVICES  INC. with reference PAYMENT",PAYMENT,272.63,,,,"AMAZON AUSTRALIA SERVICES  INC.",,,,,,,,0.00,
TRANSFER-1356004879,30-12-2024,19.87,USD,"Received money from AMAZON.COM SERVICES LLC with reference PAYMENT",PAYMENT,254.87,,,,"AMAZON.COM SERVICES LLC",,,,,,,,0.00,
TRANSFER-1356003098,30-12-2024,12.33,USD,"Received money from AMAZON MEDIA EU S.A.R.L. with reference PAYMENT",PAYMENT,235.00,,,,"AMAZON MEDIA EU S.A.R.L.",,,,,,,,0.00,
CARD-2106097800,30-12-2024,-10.00,USD,"Card transaction of USD issued by Booksirens PHILADELPHIA",,222.67,,,,,,,"Booksirens PHILADELPHIA",1033,"Yelizaveta Shulyayeva",,,0.00,
TRANSFER-1355191945,30-12-2024,0.02,USD,"Received money from AMAZON SE5097806 with reference EDI PYMNTS","EDI PYMNTS",232.67,,,,"AMAZON SE5097806",,,,,,,,0.00,
`,
			llmRes: func(t *testing.T) (map[string]int, error) {
				return map[string]int{
					"date":        1,
					"amount":      2,
					"currency":    3,
					"description": 4,
				}, nil
			},
			wantDoc: document.Document{
				Transactions: []document.Transaction{
					{
						Amount:      17.76,
						Currency:    document.USD,
						Date:        date_30122024,
						Description: "Received money from AMAZON AUSTRALIA SERVICES  INC. with reference PAYMENT",
					},
					{
						Amount:      19.87,
						Currency:    document.USD,
						Date:        date_30122024,
						Description: "Received money from AMAZON.COM SERVICES LLC with reference PAYMENT",
					},
					{
						Amount:      12.33,
						Currency:    document.USD,
						Date:        date_30122024,
						Description: "Received money from AMAZON MEDIA EU S.A.R.L. with reference PAYMENT",
					},
					{
						Amount:      -10.00,
						Currency:    document.USD,
						Date:        date_30122024,
						Description: "Card transaction of USD issued by Booksirens PHILADELPHIA",
					},
					{
						Amount:      0.02,
						Currency:    document.USD,
						Date:        date_30122024,
						Description: "Received money from AMAZON SE5097806 with reference EDI PYMNTS",
					},
				},
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockLlm := mocks.NewLlm(t)
			mockLlm.On("FindElements", mock.Anything, mock.Anything).Return(tc.llmRes(t))

			parser := parser.NewParser(mockLlm)

			reader := csv.NewReader(strings.NewReader(tc.doc))
			reader.LazyQuotes = true // Handle inconsistent quotes
			records, err := reader.ReadAll()
			require.NoError(t, err)

			gotDoc, gotErr := parser.Parse(records)
			require.ErrorIs(t, gotErr, tc.wantErr)
			if gotErr == nil {
				require.EqualValues(t, &tc.wantDoc, gotDoc)
			}
		})
	}
}
