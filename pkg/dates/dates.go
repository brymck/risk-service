package dates

import (
	"fmt"
	"time"

	sec "github.com/brymck/risk-service/genproto/brymck/securities/v1"
)

type DateComparisonResult string

const (
	EndOfDayHour                      = 17
	Before       DateComparisonResult = "Before"
	Equal        DateComparisonResult = "Equal"
	After        DateComparisonResult = "After"
)

var (
	easternTime *time.Location
)

func init() {
	var err error
	easternTime, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(fmt.Sprintf("unable to load Eastern Time information: %v", err))
	}
}

func ToProtoDate(t time.Time) *sec.Date {
	return &sec.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func Compare(d *sec.Date, year int, month time.Month, day int) DateComparisonResult {
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

func LatestBusinessDate() time.Time {
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
