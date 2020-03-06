package main

import (
	"github.com/brymck/helpers/servers"
	"github.com/brymck/helpers/services"

	rk "github.com/brymck/risk-service/genproto/brymck/risk/v1"
	sec "github.com/brymck/risk-service/genproto/brymck/securities/v1"
)

type application struct {
	securities sec.SecuritiesAPIClient
}

func main() {
	app := &application{
		securities: sec.NewSecuritiesAPIClient(services.MustConnect("securities-service")),
	}

	s := servers.NewGrpcServer()
	rk.RegisterRiskAPIServer(s.Server, app)
	s.Serve()
}
