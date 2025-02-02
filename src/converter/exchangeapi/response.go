package exchangeapi

import "encoding/json"

type ApiResponse struct {
	Date  string `json:"date"`
	Rates map[string]map[string]float64
}

func (r *ApiResponse) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	r.Rates = make(map[string]map[string]float64)

	for key, value := range raw {
		if key == "date" {
			r.Date = value.(string)
			continue
		}

		// Any other key is assumed to be a currency code
		if rates, ok := value.(map[string]interface{}); ok {
			r.Rates[key] = make(map[string]float64)
			for currency, rate := range rates {
				r.Rates[key][currency] = rate.(float64)
			}
		}
	}

	return nil
}
