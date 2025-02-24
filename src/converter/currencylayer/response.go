package currencylayer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/lazeratops/optimusdime/src/document"
)

type UnixTimestamp string

func (ut *UnixTimestamp) UnmarshalJSON(data []byte) error {
	var timestamp string
	if err := json.Unmarshal(data, &timestamp); err != nil {
		return err
	}
	*ut = UnixTimestamp(timestamp)
	return nil
}

func (ut UnixTimestamp) Time() time.Time {
	n, _ := strconv.Atoi(string(ut))
	return time.Unix(int64(n), 0)
}

type ApiResponse struct {
	Success    bool               `json:"success"`
	Historical bool               `json:"historical"`
	Date       UnixTimestamp      `json:"date"`
	Timestamp  int                `json:"timestamp"`
	Source     document.Currency  `json:"source"`
	Quotes     map[string]float64 `json:"quotes"`
}

func (r *ApiResponse) getCrossRate(sourceCurrency, targetCurrency document.Currency) (float64, error) {
	if sourceCurrency == document.USD {
		return r.getUsdTargetRate(targetCurrency)
	}
	// Get USD to source rate
	usdSourceRate, exists := r.Quotes[fmt.Sprintf("USD%s", sourceCurrency.String())]
	if !exists {
		return 0, fmt.Errorf("failed to find rate for USD to %s", sourceCurrency)
	}

	// Get USD to target rate
	usdTargetRate, err := r.getUsdTargetRate(targetCurrency)
	if err != nil {
		return 0, err
	}

	// Calculate cross rate: source -> USD -> target
	crossRate := usdTargetRate / usdSourceRate

	return crossRate, nil
}

func (r *ApiResponse) getUsdTargetRate(targetCurrency document.Currency) (float64, error) {
	usdTargetRate, exists := r.Quotes[fmt.Sprintf("USD%s", targetCurrency.String())]
	if !exists {
		return 0, fmt.Errorf("failed to find rate for USD to %s", targetCurrency)
	}
	return usdTargetRate, nil
}
