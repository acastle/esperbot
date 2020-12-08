package util

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

type ParseError struct {
	Input string
	Err   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("input could not be parsed: %s", e.Err.Error())
}
func (e *ParseError) Unwrap() error { return e.Err }
func (e *ParseError) Is(err error) bool {
	t, ok := err.(*ParseError)
	if !ok {
		return false
	}

	return t.Input == e.Input
}

// FlagsToDateRange returns a DateRange object from a slice of text flags.
// The flags can be provided in the following forms:
// 1. [] where the range is assumed to be the current week starting Sunday at 00:00:00 and ending the following Saturday at 23:59:59
// 2. ["12/20"] where the range is assumed to be a single day beginning at 00:00:00 and ending at 23:59:59
// 3. ["12/20", "to", "12/21"] where the date is assumed to be 00:00:00 on the starting date to 23:59:59 on the final date
// This method also supports dates that are split into multiple words. For example ["Dec", "20", "2020"] will be joined and parsed as a single word "Dec 20 2020"
func FlagsToDateRange(flags []string) (DateRange, error) {
	// Form #1 no dates provided
	if len(flags) == 0 {
		now := time.Now().UTC()
		return DateRange{
			Begin: BeginningOfWeek(now),
			End:   EndOfWeek(now),
		}, nil
	}

	toIdx := -1
	for i, f := range flags {
		if strings.ToLower(f) == "to" {
			toIdx = i
		}
	}

	// Form #2 singular date provided
	if toIdx == -1 {
		dateStr := strings.Join(flags, " ")
		d, err := dateparse.ParseAny(dateStr)
		if err != nil {
			return DateRange{}, &ParseError{
				Input: dateStr,
				Err:   err,
			}
		}

		d = RelativeDefaults(d)
		return DateRange{
			Begin: BeginningOfDay(d),
			End:   EndOfDay(d),
		}, nil
	}

	// Form #3 with date range
	beginStr := strings.Join(flags[:toIdx], " ")
	endStr := strings.Join(flags[toIdx+1:], " ")

	begin, err := dateparse.ParseAny(beginStr)
	if err != nil {
		return DateRange{}, &ParseError{
			Input: beginStr,
			Err:   err,
		}
	}

	end, err := dateparse.ParseAny(endStr)
	if err != nil {
		return DateRange{}, &ParseError{
			Input: endStr,
			Err:   err,
		}
	}

	begin = RelativeDefaults(begin)
	end = RelativeDefaults(end)
	return DateRange{
		Begin: BeginningOfDay(begin),
		End:   EndOfDay(end),
	}, nil
}

// DayOfWeek returns a new time with the same wall time for the specified day of the week
func DayOfWeek(t time.Time, weekday time.Weekday) time.Time {
	begin := BeginningOfWeek(t)
	year, month, day := begin.Date()
	return time.Date(year, month, day+int(weekday), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// BeginningOfDay returns a new time at 00:00:00 of the current day in the provided timezone
func BeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay returns a new time at 23:59:59+999999999 of the current day in the provided timezone
func EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day+1, 0, 0, 0, 0, t.Location()).Add(-1 * time.Nanosecond)
}

// BeginningOfWeek returns a new time at 00:00:00 on Sunday of the provided time's week and timezone
func BeginningOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	year, month, day := t.Date()
	return time.Date(year, month, day-int(weekday), 0, 0, 0, 0, t.Location())
}

// EndOfWeek returns a new time at 23:59:59+999999999 on Saturday of the provided time's week and timezone
func EndOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	year, month, day := t.Date()
	return time.Date(year, month, day+(7-int(weekday)), 0, 0, 0, 0, t.Location()).Add(-1 * time.Nanosecond)
}

// RelativeDefaults returns a new time defaulted with the current date and time. This is useful for when a
// date is supplied with no month or year provided.
func RelativeDefaults(t time.Time) time.Time {
	now := time.Now().UTC()
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	if year == 0 {
		year = now.Year()
	}

	if month < now.Month() {
		year = year + 1
	}

	return time.Date(year, month, day, hour, minute, second, t.Nanosecond(), t.Location())
}

var ErrDateRangeTooLarge = errors.New("date range must be less than one year")

func ForEachPeriod(r DateRange, do func(time.Time) error, year int, month int, day int) error {
	if r.End.Sub(r.Begin) > time.Hour*24*365 {
		return ErrDateRangeTooLarge
	}

	next := r.Begin
	for next.Before(r.End) {
		err := do(next)
		if err != nil {
			return fmt.Errorf("apply function for date: %w", err)
		}

		next = next.AddDate(year, month, day)
	}

	return nil
}

func ForEachDay(r DateRange, do func(time.Time) error) error {
	return ForEachPeriod(r, do, 0, 0, 1)
}

func ForEachWeek(r DateRange, do func(time.Time) error) error {
	return ForEachPeriod(r, do, 0, 0, 7)
}
