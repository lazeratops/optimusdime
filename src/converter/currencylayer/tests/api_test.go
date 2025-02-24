package apitest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lazeratops/optimusdime/src/converter"
	"github.com/lazeratops/optimusdime/src/converter/currencylayer"
	"github.com/lazeratops/optimusdime/src/document"

	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	t.Parallel()
	dateStr := "2025-01-01"
	date_20250101, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t.Fatalf("failed to parse date: %v", err)
	}
	dateUnix := date_20250101.Unix()

	cases := []struct {
		name           string
		apiStatusCode  int
		apiBody        string
		sourceDocument *document.Document
		wantDocument   *document.Document
		targetCurrency document.Currency
		wantErr        error
	}{
		{
			name:           "bad status code",
			targetCurrency: document.EUR,
			sourceDocument: &document.Document{
				Transactions: []document.Transaction{
					{
						Description: "transaction1",
						Currency:    document.SEK,
						Amount:      100,
						Date:        date_20250101,
					},
				},
			},
			apiStatusCode: http.StatusInternalServerError,
			wantErr:       converter.ErrFailedAPICall,
		},
		{
			name:          "success",
			apiStatusCode: http.StatusOK,
			// Updated to match CurrencyLayer API response format
			apiBody: fmt.Sprintf(`{
				"success": true,
				"terms": "https://currencylayer.com/terms",
				"privacy": "https://currencylayer.com/privacy",
				"historical": true,
				"date": "%d",
				"source": "USD",
				"quotes": {
                    "USDSEK": 11.24233239,
                    "USDEUR": 1.0
				}
			}`, dateUnix),
			sourceDocument: &document.Document{
				Transactions: []document.Transaction{
					{
						Description: "transaction1",
						Currency:    document.SEK,
						Amount:      1124.233239,
						Date:        date_20250101,
					},
				},
			},
			targetCurrency: document.EUR,
			wantDocument: &document.Document{
				Transactions: []document.Transaction{
					{
						Description: "transaction1",
						Currency:    document.EUR,
						Amount:      100,
						Date:        date_20250101,
					},
				},
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request parameters
				require.Contains(t, r.URL.String(), "access_key=some-key")
				require.Contains(t, r.URL.String(), "date="+dateStr)
				require.Contains(t, r.URL.String(), "currencies=")

				w.WriteHeader(tc.apiStatusCode)
				_, err := w.Write([]byte(tc.apiBody))
				require.NoError(t, err)
			}))

			defer testServer.Close()

			api, err := currencylayer.NewCurrencyLayer(testServer.URL, "some-key")
			require.NoError(t, err)

			gotDoc, failedDoc, gotErr := api.Convert(tc.targetCurrency, tc.sourceDocument)
			require.ErrorIs(t, gotErr, tc.wantErr)
			if gotErr == nil {
				require.EqualValues(t, tc.wantDocument, gotDoc)
				require.Empty(t, failedDoc.Transactions)
			}
		})
	}
}
