package exchangeapi

import (
	"fmt"
	"testing"
	"time"

	"github.com/lazeratops/optimusdime/src/document"
	"github.com/stretchr/testify/require"
)

func TestGetUrl(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		date           string
		targetCurrency document.Currency
		wantUrl        string
	}{
		{
			name:           "bad status code",
			date:           "2020-11-24",
			targetCurrency: document.SEK,
			wantUrl:        fmt.Sprintf("https://2020-11-24.%s/sek.json", defaultApiUrl),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parsedTime, err := time.Parse("2006-01-02", tc.date)
			require.NoError(t, err)
			api, err := NewExchangeApi("")
			require.NoError(t, err)
			gotUrl := api.getUrl(parsedTime, tc.targetCurrency)
			require.EqualValues(t, tc.wantUrl, gotUrl)
		})
	}
}
