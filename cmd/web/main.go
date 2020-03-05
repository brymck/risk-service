package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/brymck/helpers/servers"
	log "github.com/sirupsen/logrus"

	rk "github.com/brymck/risk-service/genproto/brymck/risk/v1"
	sec "github.com/brymck/risk-service/genproto/brymck/securities/v1"
)

type dateComparisonResult string

const (
	EndOfDayHour                      = 17
	Before       dateComparisonResult = "Before"
	Equal        dateComparisonResult = "Equal"
	After        dateComparisonResult = "After"
)

var (
	easternTime *time.Location
)

type application struct {
}

func calculateCovariance(xs []float64, ys []float64) float64 {
	sumOfSquares := 0.0
	count := len(xs)
	if count == 0 {
		return 0.0
	}
	for i := 0; i < count; i++ {
		sumOfSquares += xs[i] * ys[i]
	}
	return sumOfSquares / float64(count)
}

func calculateReturns(xs []float64) []float64 {
	count := len(xs)
	if count <= 1 {
		return []float64{}
	}
	results := make([]float64, len(xs)-1)
	prev := xs[0]
	for i := 1; i < count; i++ {
		x := xs[i]
		if prev == 0.0 {
			results[i-1] = 0.0
		} else {
			results[i-1] = x/prev - 1.0
		}
	}
	return results
}

func (app *application) GetCovariances(ctx context.Context, in *rk.GetCovariancesRequest) (*rk.GetCovariancesResponse, error) {
	end := latestBusinessDate()
	start := end.AddDate(-1, 0, 0)

	securityIds := in.SecurityIds
	timeSeries := make(map[uint64][]float64, len(securityIds))
	for _, securityId := range securityIds {
		log.Infof("getting prices for %d", securityId)
		entries, err := getPrices(ctx, securityId, &start, &end)
		if err != nil {
			return nil, err
		}
		log.Infof(
			"normalizing %d time series entries for %d from %s to %s",
			len(entries),
			securityId,
			isoDate(&start),
			isoDate(&end),
		)
		timeSeries[securityId] = calculateReturns(normalizeTimeSeries(entries, start, end))
	}

	log.Info("calculating covariances")
	var pairs []*rk.CovariancePair
	count := len(in.SecurityIds)
	for i := 0; i < count; i++ {
		securityId1 := securityIds[i]
		for j := i; j < count; j++ {
			securityId2 := securityIds[j]
			covariance := calculateCovariance(timeSeries[securityId1], timeSeries[securityId2])
			pair := &rk.CovariancePair{SecurityId1: securityId1, SecurityId2: securityId2, Covariance: covariance}
			pairs = append(pairs, pair)
		}
	}

	log.Info("responding with %d covariance pairs", len(pairs))
	return &rk.GetCovariancesResponse{Covariances: pairs}, nil
}

func init() {
	var err error
	easternTime, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(fmt.Sprintf("unable to load Eastern Time information: %v", err))
	}
}

func getPrices(ctx context.Context, securityId uint64, start *time.Time, end *time.Time) ([]*sec.Price, error) {
	req := &sec.GetPricesRequest{Id: securityId, StartDate: toProtoDate(start), EndDate: toProtoDate(end)}
	resp, err := securitiesApi.GetPrices(ctx, req)
	if err != nil {
		return nil, nil
	}
	return resp.Prices, nil
}

func toProtoDate(t *time.Time) *sec.Date {
	return &sec.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func compareDates(d *sec.Date, year int, month time.Month, day int) dateComparisonResult {
	year32 := int32(year)
	if d.Year < year32 {
		return Before
	} else if d.Year == year32 {
		month32 := int32(month)
		if d.Month < month32 {
			return Before
		} else if d.Month == month32 {
			day32 := int32(day)
			if d.Day < day32 {
				return Before
			} else {
				return Equal
			}
		}
	}
	return After
}

func normalizeTimeSeries(entries []*sec.Price, start time.Time, end time.Time) []float64 {
	var results []float64
	i := 0
	count := len(entries)
	var lastPrice float64
	if count == 0 {
		lastPrice = 0.0
	} else {
		lastPrice = entries[count - 1].Price
	}
	for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
		weekday := d.Weekday()
		if d == start || (weekday != time.Saturday && weekday != time.Sunday) {
			year, month, day := d.Date()
			if i >= count {
				results = append(results, lastPrice)
			} else {
			loop:
				for {
					entry := entries[i]
					result := compareDates(entry.Date, year, month, day)
					switch result {
					case Equal:
						results = append(results, entry.Price)
						i++
						break loop
					case Before:
						i++
						if i >= count {
							break loop
						}
					case After:
						break loop
					}
				}
			}
		}
	}
	return results
}

func latestBusinessDate() time.Time {
	date := time.Now().In(easternTime)
	switch date.Weekday() {
	case time.Monday:
		if date.Hour() < EndOfDayHour {
			date = date.AddDate(0, 0, -3)
		}
	case time.Sunday:
		date = date.AddDate(0, 0, -2)
	case time.Saturday:
		date = date.AddDate(0, 0, -1)
	default:
		if date.Hour() < EndOfDayHour {
			date = date.AddDate(0, 0, -1)
		}
	}
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func isoDate(t *time.Time) string {
	return t.Format("2006-01-02")
}

func (app *application) GetRisk(ctx context.Context, in *rk.GetRiskRequest) (*rk.GetRiskResponse, error) {
	end := latestBusinessDate()
	start := end.AddDate(-1, 0, 0)
	log.Infof("getting prices for %d", in.SecurityId)
	entries, err := getPrices(ctx, in.SecurityId, &start, &end)
	if err != nil {
		return nil, err
	}

	log.Infof(
		"normalizing %d time series entries for %d from %s to %s",
		len(entries),
		in.SecurityId,
		isoDate(&start),
		isoDate(&end),
	)
	normalized := normalizeTimeSeries(entries, start, end)

	count := len(normalized)
	log.Infof("calculating variance of %d normalized time series entries", count)
	sumOfSquares := 0.0
	previousPrice := 0.0
	for _, price := range normalized {
		if previousPrice != 0.0 {
			sumOfSquares += math.Pow(price/previousPrice-1.0, 2.0)
		}
		previousPrice = price
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
