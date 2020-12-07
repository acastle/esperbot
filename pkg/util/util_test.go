package util

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestFlagsToDateRange(t *testing.T) {
	now := time.Now().UTC()

	cases := []struct {
		name     string
		flags    []string
		expRange DateRange
		expErr   error
	}{
		{
			"with no flags",
			[]string{},
			DateRange{
				BeginningOfWeek(now),
				EndOfWeek(now),
			},
			nil,
		},
		{
			"with singular date '12/20' fmt",
			[]string{"12/20"},
			DateRange{
				time.Date(now.Year(), 12, 20, 0, 0, 0, 0, time.UTC),
				time.Date(now.Year(), 12, 21, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			},
			nil,
		},
		{
			"with singular date '12.20' fmt",
			[]string{"12.20"},
			DateRange{
				time.Date(now.Year(), 12, 20, 0, 0, 0, 0, time.UTC),
				time.Date(now.Year(), 12, 21, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			},
			nil,
		},
		{
			"with singular date 'dec 20' fmt",
			[]string{"dec", "20"},
			DateRange{
				time.Date(now.Year(), 12, 20, 0, 0, 0, 0, time.UTC),
				time.Date(now.Year(), 12, 21, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			},
			nil,
		},
		{
			"with singular date 'dec 20 2010' fmt",
			[]string{"dec", "20", "2010"},
			DateRange{
				time.Date(2010, 12, 20, 0, 0, 0, 0, time.UTC),
				time.Date(2010, 12, 21, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			},
			nil,
		},
		{
			"with date range '12/20 to 12/25' fmt",
			[]string{"12/20", "to", "12/25"},
			DateRange{
				time.Date(now.Year(), 12, 20, 0, 0, 0, 0, time.UTC),
				time.Date(now.Year(), 12, 26, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			},
			nil,
		},
		{
			"with date range 'dec 20 to jan 1' fmt",
			[]string{"dec", "20", "to", "jan", "1"},
			DateRange{
				time.Date(now.Year(), 12, 20, 0, 0, 0, 0, time.UTC),
				time.Date(now.Year()+1, 1, 2, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			},
			nil,
		},
		{
			"err when range begins with to",
			[]string{"to", "jan", "1"},
			DateRange{},
			&ParseError{Input: ""},
		},
		{
			"err when range ends with to",
			[]string{"jan", "1", "to"},
			DateRange{},
			&ParseError{Input: ""},
		},
		{
			"err when only 'to'",
			[]string{"to"},
			DateRange{},
			&ParseError{Input: ""},
		},
		{
			"err invalid single date",
			[]string{"12231/123131"},
			DateRange{},
			&ParseError{Input: "12231/123131"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, err := FlagsToDateRange(c.flags)
			if !errors.Is(err, c.expErr) {
				t.Error("did not return the expected error")
				return
			}

			if !reflect.DeepEqual(r, c.expRange) {
				t.Errorf("did not return the correct date range, wanted '{Begin: %s, End: %s}', got '{Begin: %s, End: %s}'", c.expRange.Begin, c.expRange.End, r.Begin, r.End)
			}
		})
	}

}

func TestBeginningOfDay(t *testing.T) {
	cases := []struct {
		name    string
		time    time.Time
		expTime time.Time
	}{
		{
			"beginning of day",
			time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			"midday",
			time.Date(2020, 12, 25, 12, 0, 0, 0, time.UTC),
			time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			"end of day",
			time.Date(2020, 12, 26, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := BeginningOfDay(c.time)
			if !reflect.DeepEqual(result, c.expTime) {
				t.Errorf("incorrect date returned, wanted '%s', got '%s'", c.expTime, result)
			}
		})
	}
}

func TestEndOfDay(t *testing.T) {
	cases := []struct {
		name    string
		time    time.Time
		expTime time.Time
	}{
		{
			"beginning of day",
			time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 12, 26, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
		},
		{
			"midday",
			time.Date(2020, 12, 25, 12, 0, 0, 0, time.UTC),
			time.Date(2020, 12, 26, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
		},
		{
			"end of day",
			time.Date(2020, 12, 26, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			time.Date(2020, 12, 26, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := EndOfDay(c.time)
			if !reflect.DeepEqual(result, c.expTime) {
				t.Errorf("incorrect date returned, wanted '%s', got '%s'", c.expTime, result)
			}
		})
	}
}

func TestBeginningOfWeek(t *testing.T) {
	loc, _ := time.LoadLocation("America/Detroit")
	cases := []struct {
		name    string
		time    time.Time
		expTime time.Time
	}{
		{
			"standard week midweek",
			time.Date(2020, 12, 26, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 12, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			"standard week beginning of the week",
			time.Date(2020, 12, 20, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 12, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			"standard week end of the week",
			time.Date(2020, 12, 27, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			time.Date(2020, 12, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			"week with dst start",
			time.Date(2020, 3, 9, 0, 0, 0, 0, loc),
			time.Date(2020, 3, 8, 0, 0, 0, 0, loc),
		},
		{
			"week with dst end",
			time.Date(2020, 11, 2, 0, 0, 0, 0, loc),
			time.Date(2020, 11, 1, 0, 0, 0, 0, loc),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := BeginningOfWeek(c.time)
			if !reflect.DeepEqual(result, c.expTime) {
				t.Errorf("incorrect date returned, wanted '%s', got '%s'", c.expTime, result)
			}
		})
	}
}

func TestEndOfWeek(t *testing.T) {
	loc, _ := time.LoadLocation("America/Detroit")
	cases := []struct {
		name    string
		time    time.Time
		expTime time.Time
	}{
		{
			"standard week midweek",
			time.Date(2020, 12, 26, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 12, 27, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
		},
		{
			"standard week beginning",
			time.Date(2020, 12, 20, 0, 0, 0, 0, time.UTC),
			time.Date(2020, 12, 27, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
		},
		{
			"standard week end",
			time.Date(2020, 12, 27, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
			time.Date(2020, 12, 27, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond),
		},
		{
			"week with dst start",
			time.Date(2020, 3, 9, 0, 0, 0, 0, loc),
			time.Date(2020, 3, 15, 0, 0, 0, 0, loc).Add(-1 * time.Nanosecond),
		},
		{
			"week with dst end",
			time.Date(2020, 11, 2, 0, 0, 0, 0, loc),
			time.Date(2020, 11, 8, 0, 0, 0, 0, loc).Add(-1 * time.Nanosecond),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := EndOfWeek(c.time)
			if !reflect.DeepEqual(result, c.expTime) {
				t.Errorf("incorrect date returned, wanted '%s', got '%s'", c.expTime, result)
			}
		})
	}
}

func TestRelativeDefaults(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		name    string
		time    time.Time
		expTime time.Time
	}{
		{
			"default year",
			time.Date(0, 12, 20, 0, 0, 0, 0, time.UTC),
			time.Date(now.Year(), 12, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			"next year if month is past",
			time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"provided values are unchanged",
			time.Date(2010, 12, 20, 12, 45, 50, 1234, time.UTC),
			time.Date(2010, 12, 20, 12, 45, 50, 1234, time.UTC),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := RelativeDefaults(c.time)
			if !reflect.DeepEqual(result, c.expTime) {
				t.Errorf("incorrect date returned, wanted '%s', got '%s'", c.expTime, result)
			}
		})
	}
}
