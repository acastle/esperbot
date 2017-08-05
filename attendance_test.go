package main

import (
	"testing"
	"time"
)

func TestMissRaidResult_String(t *testing.T) {
	result := MissRaidResult{
		Name: "SomeName",
		Dates: []time.Time{
			time.Date(2017, 1, 1, 0, 0,0,0, time.UTC),
			time.Date(2017, 2, 2, 0, 0,0,0, time.UTC),
			time.Date(2017, 2, 20, 0, 0,0,0, time.UTC),
		},

	}

	output := result.String()
	if (output != "Marked **SomeName** out on Jan 1, Feb 2, and Feb 20") {
		t.Fail()
	}
}