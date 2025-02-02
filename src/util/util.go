package util

import (
	"fmt"
	"time"
)

func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",                // YYYY-MM-DD
		"02-01-2006",                // DD-MM-YYYY
		"01/02/2006",                // MM/DD/YYYY
		"2006/01/02",                // YYYY/MM/DD
		"02 Jan 2006",               // DD Mon YYYY
		"2006-01-02T15:04:05Z07:00", // ISO8601
	}

	var lastErr error
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse date '%s': %w", dateStr, lastErr)
}
