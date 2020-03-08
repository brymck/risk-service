package dates

import (
	"time"

	dt "github.com/brymck/genproto/brymck/dates/v1"
	rk "github.com/brymck/genproto/brymck/risk/v1"
)

type DateComparisonResult string
type Frequency string

const (
	Before  DateComparisonResult = "Before"
	Equal   DateComparisonResult = "Equal"
	After   DateComparisonResult = "After"
	Daily   Frequency            = "Daily"
	Monthly Frequency            = "Monthly"
)

func ToProtoDate(t time.Time) *dt.Date {
	return &dt.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func Compare(d *dt.Date, year int, month time.Month, day int) DateComparisonResult {
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

func ToFrequency(freq rk.Frequency) Frequency {
	switch freq.String() {
	case "FREQUENCY_MONTHLY":
		return Monthly
	default:
		return Daily
	}
}
