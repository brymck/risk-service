package main

import (
	"time"

	sec "github.com/brymck/genproto/brymck/securities/v1"

	"github.com/brymck/risk-service/pkg/dates"
)

func calculateCovariance(xs []float64, ys []float64) float64 {
	sumOfSquares := 0.0
	count := len(xs)
	if count == 0 {
		return 0.0
	}
	for i := 0; i < count; i++ {
		sumOfSquares += xs[i] * ys[i]
	}
	annualization := float64(count)
	return sumOfSquares / float64(count) * annualization * 10000.0
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
		prev = x
	}
	return results
}

func normalizeDates(start time.Time, end time.Time, frequency dates.Frequency) []time.Time {
	var results []time.Time
	switch frequency {
	case dates.Daily:
		for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
			weekday := d.Weekday()
			if d == start || (weekday != time.Saturday && weekday != time.Sunday) {
				results = append(results, d)
			}
		}
	case dates.Monthly:
		for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
			if d == start {
				results = append(results, d)
			} else {
				if d.AddDate(0, 0, 1).Day() == 1 {
					weekday := d.Weekday()
					switch weekday {
					case time.Saturday:
						results = append(results, d.AddDate(0, 0, -1))
					case time.Sunday:
						results = append(results, d.AddDate(0, 0, -2))
					default:
						results = append(results, d)
					}
				}
			}
		}
	}
	return results
}

func normalizeTimeSeries(entries []*sec.Price, normalizedDates []time.Time) []float64 {
	results := make([]float64, len(normalizedDates))
	i := 0
	j := 0
	count := len(entries)
	var lastPrice float64
	if count == 0 {
		lastPrice = 0.0
	} else {
		lastPrice = entries[count-1].Price
	}
	start := normalizedDates[0]
	end := normalizedDates[len(normalizedDates)-1]
	for d := start; d.After(end) == false; d = d.AddDate(0, 0, 1) {
		if d == normalizedDates[j] {
			year, month, day := d.Date()
			if i >= count {
				results[j] = lastPrice
			} else {
			loop:
				for {
					entry := entries[i]
					result := dates.Compare(entry.Date, year, month, day)
					switch result {
					case dates.Equal:
						results[j] = entry.Price
						i++
						break loop
					case dates.Before:
						i++
						if i >= count {
							results[j] = results[j-1]
							break loop
						}
					case dates.After:
						results[j] = results[j-1]
						break loop
					}
				}
			}
			j++
		}
	}
	return results
}
