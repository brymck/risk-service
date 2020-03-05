package main

import (
	"context"
	"math"

	"github.com/brymck/helpers/servers"
	_ "github.com/go-sql-driver/mysql"

	rk "github.com/brymck/risk-service/genproto/brymck/risk/v1"
	sec "github.com/brymck/risk-service/genproto/brymck/securities/v1"
)

type application struct {
}

func (app *application) GetRisk(ctx context.Context, in *rk.GetRiskRequest) (*rk.GetRiskResponse, error) {
	startDate := &sec.Date{Year: 2020, Month: 1, Day: 1}
	endDate := &sec.Date{Year: 2020, Month: 3, Day: 4}
	req := &sec.GetPricesRequest{Id: in.GetSecurityId(), StartDate: startDate, EndDate: endDate}
	resp, err := securitiesApi.GetPrices(ctx, req)
	if err != nil {
		return nil, err
	}
	count := 0
	sumOfSquares := 0.0
	previousPrice := 0.0
	for _, item := range resp.Prices {
		if previousPrice != 0.0 {
			sumOfSquares += item.Price / previousPrice - 1.0
			count++
		}
		if item.Price != 0.0 {
			previousPrice = item.Price
		}
	}
	risk := math.Sqrt(sumOfSquares / float64(count))
	return &rk.GetRiskResponse{Risk: risk}, nil
}

func main() {
	app := &application{}

	s := servers.NewGrpcServer()
	rk.RegisterRiskAPIServer(s.Server, app)
	s.Serve()
}
