package main

import (
	"testing"
	"time"
)

func TestMissRaidResult_String(t *testing.T) {
	result := missRaidResult{
		Name: "SomeName",
		Dates: []time.Time{
			time.Date(2017, 1, 1, 0, 0,0,0, time.UTC),
			time.Date(2017, 2, 2, 0, 0,0,0, time.UTC),
			time.Date(2017, 2, 20, 0, 0,0,0, time.UTC),
		},

	}

	output := result.String()
	if (output != "Marked **SomeName** out on Jan 1, Feb 2 and Feb 20") {
		t.Fail()
	}
}

func TestParsePeriod_ToLargeOfARange(t *testing.T) {
	_, err := parsePeriod("2017-08-07/2017-10-13")
	if err != ErrPeriodTooLarge {
		t.Fail()
	}
}

func TestParsePeriod(t *testing.T) {
	dates, err := parsePeriod("2017-08-07/2017-08-13")
	if err != nil {
		t.Fail()
	}

	if len(dates) != 2 {
		t.Fail()
	}

	if dates[0].Unix() != 1502150400 {
		t.Fail()
	}

	if dates[1].Unix() != 1502236800 {
		t.Fail()
	}
}

func TestParseDates(t *testing.T) {
	dates, err := parseDates([]string{"2017-08-08","2017-08-09"})
	if err != nil {
		t.Fail()
	}

	if len(dates) != 2 {
		t.Fail()
	}

	if dates[0].Unix() != 1502150400 {
		t.Fail()
	}

	if dates[1].Unix() != 1502236800 {
		t.Fail()
	}
}