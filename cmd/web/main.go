package main

import (
	"time"

	"github.com/allegro/bigcache"
	"github.com/brymck/helpers/servers"
	"github.com/brymck/helpers/services"

	rk "github.com/brymck/genproto/brymck/risk/v1"
	sec "github.com/brymck/genproto/brymck/securities/v1"
)

type application struct {
	cache      *bigcache.BigCache
	securities sec.SecuritiesAPIClient
}

func main() {
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	if err != nil {
		panic(err)
	}
	app := &application{
		cache:      cache,
		securities: sec.NewSecuritiesAPIClient(services.MustConnect("securities-service")),
	}

	s := servers.NewGrpcServer()
	rk.RegisterRiskAPIServer(s.Server, app)
	s.Serve()
}
