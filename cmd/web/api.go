package main

import (
	sec "github.com/brymck/risk-service/genproto/brymck/securities/v1"
	"github.com/brymck/risk-service/pkg/services"
)

var (
	securitiesApi sec.SecuritiesAPIClient
)

func init() {
	conn := services.NewService("securities-service").MustConnect()
	securitiesApi = sec.NewSecuritiesAPIClient(conn)
}
