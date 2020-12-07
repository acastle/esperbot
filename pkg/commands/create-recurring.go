package commands

import "time"

type CreateRecurring struct {
	Name string
	Days []time.Weekday
}
