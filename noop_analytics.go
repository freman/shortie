package main

import (
	"net/http"
)

type noopAnalytics struct {
}

func (m *noopAnalytics) Setup(c *shortieConfiguration) error {
	return nil
}

func (m *noopAnalytics) Record(r *http.Request, target, short, alias string) (*http.Cookie, error) {
	return nil, nil
}

func init() {
	RegisterMetricsInterface("noop_analytics", func() MetricsInterface {
		return &noopAnalytics{}
	})
}
