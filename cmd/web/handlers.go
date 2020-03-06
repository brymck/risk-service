package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/brymck/helpers/dates"
	log "github.com/sirupsen/logrus"

	dt "github.com/brymck/risk-service/genproto/brymck/dates/v1"
	rk "github.com/brymck/risk-service/genproto/brymck/risk/v1"
	dates2 "github.com/brymck/risk-service/pkg/dates"
)

func getSecurityIdsKey(ids []uint64) [32]byte {
	var builder strings.Builder
	for _, id := range ids {
		builder.WriteString(strconv.FormatUint(id, 16))
		builder.WriteString("-")
	}
	return sha256.Sum256([]byte(builder.String()))
}

func (app *application) GetCovariances(ctx context.Context, in *rk.GetCovariancesRequest) (*rk.GetCovariancesResponse, error) {
	response := &rk.GetCovariancesResponse{}

	end, err := dates.LatestBusinessDate()
	if err != nil {
		return nil, err
	}
	start := end.AddDate(-1, 0, 0)
	securityIdsKey := getSecurityIdsKey(in.SecurityIds)
	endDateText := dates.IsoFormat(end)
	key := fmt.Sprintf("covariances-%d-%s", securityIdsKey, endDateText)

	if err := app.getCache(key, response); err == nil {
		return response, nil
	}

	freq := dates2.ToFrequency(in.Frequency)
	priceDates := normalizeDates(start, end, freq)

	securityIds := in.SecurityIds
	timeSeries := make(map[uint64][]float64, len(securityIds))
	for _, securityId := range securityIds {
		entries, err := app.getPrices(ctx, securityId, start, end)
		if err != nil {
			return nil, err
		}
		timeSeries[securityId] = calculateReturns(normalizeTimeSeries(entries, priceDates))
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

	response = &rk.GetCovariancesResponse{Covariances: pairs}
	_ = app.setCache(key, response)
	return response, nil
}

func (app *application) GetRisk(ctx context.Context, in *rk.GetRiskRequest) (*rk.GetRiskResponse, error) {
	response := &rk.GetRiskResponse{}

	end, err := dates.LatestBusinessDate()
	if err != nil {
		return nil, err
	}
	start := end.AddDate(-1, 0, 0)
	endDateText := dates.IsoFormat(end)
	key := fmt.Sprintf("risk-%d-%s", in.SecurityId, endDateText)

	if err := app.getCache(key, response); err == nil {
		return response, nil
	}

	entries, err := app.getPrices(ctx, in.SecurityId, start, end)
	if err != nil {
		return nil, err
	}

	freq := dates2.ToFrequency(in.Frequency)
	priceDates := normalizeDates(start, end, freq)
	timeSeries := calculateReturns(normalizeTimeSeries(entries, priceDates))
	covariance := calculateCovariance(timeSeries, timeSeries)
	risk := math.Sqrt(covariance)

	response = &rk.GetRiskResponse{Risk: risk}
	_ = app.setCache(key, response)
	return response, nil
}

func (app *application) GetReturnTimeSeries(ctx context.Context, in *rk.GetReturnTimeSeriesRequest) (*rk.GetReturnTimeSeriesResponse, error) {
	response := &rk.GetReturnTimeSeriesResponse{}

	end, err := dates.LatestBusinessDate()
	if err != nil {
		return nil, err
	}
	start := end.AddDate(-1, 0, 0)
	endDateText := dates.IsoFormat(end)
	key := fmt.Sprintf("returns-%d-%s", in.SecurityId, endDateText)

	if err := app.getCache(key, response); err == nil {
		return response, nil
	}

	priceEntries, err := app.getPrices(ctx, in.SecurityId, start, end)
	if err != nil {
		return nil, err
	}

	freq := dates2.ToFrequency(in.Frequency)
	priceDates := normalizeDates(start, end, freq)
	returnValues := calculateReturns(normalizeTimeSeries(priceEntries, priceDates))
	entries := make([]*rk.ReturnTimeSeriesEntry, len(returnValues))
	for i, r := range returnValues {
		year, month, day := priceDates[i+1].Date()
		date := &dt.Date{Year: int32(year), Month: int32(month), Day: int32(day)}
		entries[i] = &rk.ReturnTimeSeriesEntry{Date: date, Return: r}
	}

	response = &rk.GetReturnTimeSeriesResponse{Entries: entries}
	_ = app.setCache(key, response)
	return response, nil
}
