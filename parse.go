package main

import (
	"errors"
	"fmt"
	"github.com/mlabouardy/dialogflow-go-client/models"
	"strings"
	"time"
)

var 
(
  ErrNoHandlerConfigured = errors.New("No handler defined for this intent")
  ErrNoValidDatesSpecified = errors.New("No valid dates were specified in this request")
)

func DispatchActions(user User, resp models.QueryResponse) ([]Result, error) {
	var err error
	var result Result
	action := resp.Result.Action
	parts := strings.Split(action, ".")

	var dates []time.Time
	if parts[0] == "whois" {
		members, err := ParseMemberParam(user, resp)
		if err != nil {
			return nil, err
		}

		return []Result{&whoisResult{members[0]}}, nil
	}

	if parts[0] == "date" {
		dates, err = ParseDateParam(resp)
	} else if parts[0] == "period" {
		dates, err = ParsePeriodParam(resp)
	}

	if err != nil {
		return nil, err
	}

	if parts[1] == "in" || parts[1] == "out" {
		members, err := ParseMemberParam(user, resp)
		if err != nil {
			return nil, err
		}

		if parts[1] == "in" {
			result, err = In(members, dates)
		} else if parts[1] == "late" {
			result, err = Late(members, dates)
		} else {
			result, err = Out(members, dates)
		}

		if err != nil {
			return nil, err
		}

		return []Result{result}, nil
	} else if parts[1] == "query" {
		return Query(dates)
	}

	return nil, ErrNoHandlerConfigured
}

func ParseMemberParam(fallback User, resp models.QueryResponse) ([]User, error) {
	vals := resp.Result.Parameters["member"]
	var users []User
	for _, val := range vals {
		users = append(users, User(val))
	}

	if len(users) == 0 {
		users = append(users, fallback)
	}

	return users, nil
}

func ParseDateParam(resp models.QueryResponse) ([]time.Time, error) {
	var dates []time.Time
	vals := resp.Result.Parameters["date"]
	if len(vals) > 0 {
		for _, s := range vals {
			date, err := parseDate(s)
			if err != nil {
				return nil, err
			}

			dates = append(dates, date)
		}
    
    return dates, nil
	}

	return nil, ErrNoValidDatesSpecified
}

func parseDate(val string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", val)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}

func ParsePeriodParam(resp models.QueryResponse) ([]time.Time, error) {
	var allDates []time.Time
	vals := resp.Result.Parameters["period"]
	for _, val := range vals {
		dates, err := parsePeriod(val)
		if err != nil {
			return nil, err
		}

		allDates = append(allDates, dates...)
	}

  if len(allDates) > 0 {
    return allDates, nil
  }
  
	return nil, ErrNoValidDatesSpecified
}

func parsePeriod(val string) ([]time.Time, error) {
	split := strings.Split(val, "/")
	if len(split) != 2 {
		return nil, ErrInvalidPeriod
	}

	start, err := time.Parse("2006-01-02", split[0])
	if err != nil {
		return nil, err
	}

	end, err := time.Parse("2006-01-02", split[1])
	if err != nil {
		return nil, err
	}

	if end.Sub(start).Hours() > (4 * 7 * 24) {
		return nil, ErrPeriodTooLarge
	}

	var dates []time.Time
	current := start
	for current.Before(end) {
		if IsValidRaidDay(current.Weekday()) {
			dates = append(dates, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	return dates, nil
}

type whoisResult struct {
	Alias User
}

func (r *whoisResult) String() string {
	return fmt.Sprintf("I think it is **%v**...", r.Alias)
}
