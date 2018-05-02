package main

import "net/http"

var metricsInterfaces = map[string]func() MetricsInterface{}

type MetricsInterface interface {
	Setup(*shortieConfiguration) error
	Record(*http.Request, string, string, string) (*http.Cookie, error)
}

func RegisterMetricsInterface(name string, f func() MetricsInterface) {
	metricsInterfaces[name] = f
}

func GetMetricsInterface(name string) (metricsInterface MetricsInterface) {
	if metricsInterfacef, ok := metricsInterfaces[name]; ok {
		metricsInterface = metricsInterfacef()
	}
	return
}
