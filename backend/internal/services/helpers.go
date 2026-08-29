package services

import (
	"fmt"
	"time"
)

func parseDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q (expected YYYY-MM-DD)", s)
	}
	return &t, nil
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
