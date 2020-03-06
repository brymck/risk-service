package main

import (
	"context"
	"math"

	dt "github.com/brymck/helpers/dates"
	log "github.com/sirupsen/logrus"

	rk "github.com/brymck/risk-service/genproto/brymck/risk/v1"
	"github.com/brymck/risk-service/pkg/dates"
)

func (app *application) GetCovariances(ctx context.Context, in *rk.GetCovariancesRequest) (*rk.GetCovariancesResponse, error) {
	end := dates.LatestBusinessDate()
	start := end.AddDate(-1, 0, 0)

	securityIds := in.SecurityIds
	timeSeries := make(map[uint64][]float64, len(securityIds))
	for _, securityId := range securityIds {
		log.Debugf("getting prices for %d", securityId)
		entries, err := app.getPrices(ctx, securityId, start, end)
		if err != nil {
			return nil, err
		}
		log.Debugf(
			"normalizing %d time series entries for %d from %s to %s",
			len(entries),
			securityId,
			dt.IsoFormat(start),
			dt.IsoFormat(end),
		)
		timeSeries[securityId] = calculateReturns(normalizeTimeSeries(entries, start, end))
	}

	log.Debugf("calculating covariances")
	count := len(in.SecurityIds)
	pairs := make([]*rk.CovariancePair, count*(count+1)/2)
	k := 0
	for i := 0; i < count; i++ {
		securityId1 := securityIds[i]
		for j := i; j < count; j++ {
			securityId2 := securityIds[j]
			covariance := calculateCovariance(timeSeries[securityId1], timeSeries[securityId2])
			pair := &rk.CovariancePair{SecurityId1: securityId1, SecurityId2: securityId2, Covariance: covariance}
			pairs[k] = pair
			k++
		}
	}

	log.Debugf("responding with %d covariance pairs", len(pairs))
	return &rk.GetCovariancesResponse{Covariances: pairs}, nil
}

func (app *application) GetRisk(ctx context.Context, in *rk.GetRiskRequest) (*rk.GetRiskResponse, error) {
	end := dates.LatestBusinessDate()
	start := end.AddDate(-1, 0, 0)
	log.Debugf("getting prices for %d", in.SecurityId)
	entries, err := app.getPrices(ctx, in.SecurityId, start, end)
	if err != nil {
		return nil, err
	}

	log.Debugf(
		"normalizing %d time series entries for %d from %s to %s",
		len(entries),
		in.SecurityId,
		dt.IsoFormat(start),
		dt.IsoFormat(end),
	)
	timeSeries := calculateReturns(normalizeTimeSeries(entries, start, end))
	covariance := calculateCovariance(timeSeries, timeSeries)
	risk := math.Sqrt(covariance)
	return &rk.GetRiskResponse{Risk: risk}, nil
}
