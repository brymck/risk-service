package main

import (
	"context"
	"time"

	sec "github.com/brymck/genproto/brymck/securities/v1"
	"github.com/brymck/risk-service/pkg/dates"
)

func (app *application) getPrices(ctx context.Context, securityId uint64, start time.Time, end time.Time) ([]*sec.Price, error) {
	req := &sec.GetPricesRequest{Id: securityId, StartDate: dates.ToProtoDate(start), EndDate: dates.ToProtoDate(end)}
	resp, err := app.securities.GetPrices(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Prices, nil
}
