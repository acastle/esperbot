package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidRaidDay   = errors.New("Not a valid raid day")
	ErrNoDatesSpecified = errors.New("Need to specifiy dates")
	ErrInvalidPeriod    = errors.New("Invalid time period given")
	ErrPeriodTooLarge   = errors.New("Time period is larger than a month")
)

func IsValidRaidDay(weekday time.Weekday) bool {
	return weekday == 2 || weekday == 3
}

func getSetKey(time time.Time) string {
	year, month, day := time.Date()
	return fmt.Sprintf("attendance:%d:%d:%d", year, month, day)
}

type missRaidResult struct {
	Users []User
	Dates []time.Time
}

type cancelMissResult struct {
	Users []User
	Dates []time.Time
}

func Out(users []User, dates []time.Time) (Result, error) {
	result := &missRaidResult{
		users,
		dates,
	}
	err := runForAll(users, dates, out)
	return result, err
}

func In(users []User, dates []time.Time) (Result, error) {
	result := &cancelMissResult{
		users,
		dates,
	}
	err := runForAll(users, dates, in)
	return result, err
}

func runForAll(users []User, dates []time.Time, action func(user User, date time.Time) error) error {
	for _, user := range users {
		for _, date := range dates {
			err := action(user, date)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func out(user User, date time.Time) error {
	if !IsValidRaidDay(date.Weekday()) {
		return ErrInvalidRaidDay
	}

	key := getSetKey(date)
	result := Redis.SAdd(key, string(user))
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

func in(user User, date time.Time) error {
	if !IsValidRaidDay(date.Weekday()) {
		return ErrInvalidRaidDay
	}

	key := getSetKey(date)
	result := Redis.SRem(key, string(user))
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

type queryResult struct {
	Date    time.Time
	Members []string
}

func Query(dates []time.Time) ([]Result, error) {
	results := []Result{}

	for _, d := range dates {
		if !IsValidRaidDay(d.Weekday()) {
			return results, ErrInvalidRaidDay
		}

		key := getSetKey(d)
		result := Redis.SMembers(key)
		results = append(results, &queryResult{d, result.Val()})
	}

	return results, nil
}

func (r *missRaidResult) String() string {
	return fmt.Sprintf("Marked %v out on %v", formatUsers(r.Users), formatDates(r.Dates))
}

func (r *cancelMissResult) String() string {
	return fmt.Sprintf("Marked %v in on %v", formatUsers(r.Users), formatDates(r.Dates))
}

func formatUsers(users []User) string {
	var names []string
	for _, user := range users {
		names = append(names, fmt.Sprintf("**%v**", user))
	}

	if len(users) > 1 {
		commaString := strings.Join(names[:len(names)-1], ", ")
		return fmt.Sprintf("%s and %s", commaString, names[len(names)-1])
	} else {
		return names[0]
	}
}

func formatDates(dates []time.Time) string {
	formatted := []string{}
	for _, d := range dates {
		formatted = append(formatted, d.Format("Jan 2"))
	}

	if len(dates) > 1 {
		commaString := strings.Join(formatted[:len(formatted)-1], ", ")
		return fmt.Sprintf("%s and %s", commaString, formatted[len(formatted)-1])
	} else {
		return formatted[0]
	}
}

func (r *queryResult) String() string {
	year, month, day := r.Date.Date()
	var membersList string
	if len(r.Members) == 0 {
		membersList = "No one is out :thumbsup:"
	} else {
		membersList = strings.Join(r.Members, "\n  ")
	}

	return fmt.Sprintf("**Raiders out for %d/%d/%d**\n  %s", month, day, year, membersList)
}

type Result interface {
	String() string
}
