package dates

import (
	"time"

	sec "github.com/brymck/risk-service/genproto/brymck/securities/v1"
)

type DateComparisonResult string

const (
	Before DateComparisonResult = "Before"
	Equal  DateComparisonResult = "Equal"
	After  DateComparisonResult = "After"
)

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
