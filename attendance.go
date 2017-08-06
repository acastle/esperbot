package main

import (
  "time"
  "errors"
  "fmt"
  "strings"
  "github.com/acastle/apiai-go"
  "github.com/bwmarrin/discordgo"
  "github.com/go-redis/redis"
)

var (
  ErrInvalidRaidDay = errors.New("Not a valid raid day")
  ErrNoDatesSpecified = errors.New("Need to specifiy dates")
  ErrInvalidPeriod = errors.New("Invalid time period given")
  ErrPeriodTooLarge = errors.New("Time period is larger than a month")
)

func IsValidRaidDay(weekday time.Weekday) bool {
  return weekday == 2 || weekday == 3
}

func getSetKey(time time.Time) string {
  year,month,day := time.Date()
  return fmt.Sprintf("attendance:%d:%d:%d", year, month, day)
}

func parseDates (dateStrings []string) ([]time.Time, error) {
  var dates []time.Time
  if len(dateStrings) > 0 {
    for _,s := range dateStrings {
      date, err := time.Parse("2006-01-02", s)
      if err != nil {
        return []time.Time{}, err
      }

      dates = append(dates, date)
    }
  }

  return dates, nil
}

func parsePeriod (period string) ([]time.Time, error) {
  var dates []time.Time
  split := strings.Split(period, "/")
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

  current:=start
  for current.Before(end) {
    if IsValidRaidDay(current.Weekday()) {
      dates = append(dates, current)
    }
    current=current.AddDate(0,0,1)
  }

  return dates, nil
}

func getDates(resp apiaigo.ResponseStruct) ([]time.Time, error){
  dateStrings := resp.Result.Parameters["date"].Values
  period := resp.Result.Parameters["period"].Value

  var dates []time.Time
  var err error
  if len(dateStrings) > 0 {
    dates, err = parseDates(dateStrings)
  } else if (period != "") {
    dates, err = parsePeriod(period)
  }

  if err != nil {
    return nil, err
  }

  if len(dates) > 0 {
    return dates, nil
  }

  return nil, ErrNoDatesSpecified
}

type listParam struct {
  data []interface{}
}

type missRaidResult struct {
  Name string
  Dates []time.Time
}

type cancelMissResult struct {
  Name string
  Dates []time.Time
}

func MissRaid(user *discordgo.User, resp apiaigo.ResponseStruct, cancel bool) ([]Result, error) {
  members := resp.Result.Parameters["member"].Values
  if len(members) == 0 {
    members = []string{user.Username}
  }

  dates, err := getDates(resp)
  if err != nil {
    return []Result{}, err
  }

  results := []Result{}
  for _,n := range members {
    val := strings.Title(n)

    for _,d := range dates {
      if !IsValidRaidDay(d.Weekday()) {
        continue
      }

      key := getSetKey(d)
      var result *redis.IntCmd
      if cancel {
        result = Redis.SRem(key, val)
        results = append(results, &cancelMissResult{val, dates})
      } else {
        result = Redis.SAdd(key, val)
        results = append(results, &missRaidResult{val, dates})
      }
      if result.Err() != nil {
        return []Result{},result.Err()
      }
    }
  }

  return results, nil
}

type queryResult struct {
  Date time.Time
  Members []string
}

func Query(resp apiaigo.ResponseStruct) ([]Result, error){
  results := []Result{}
  dates, err := getDates(resp)
  if err != nil {
    return results, err
  }

  for _,d := range dates {
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
  return fmt.Sprintf("Marked **%v** out on %v", r.Name, formatDates(r.Dates))
}

func (r *cancelMissResult) String() string {
  return fmt.Sprintf("Marked **%v** in on %v", r.Name, formatDates(r.Dates))
}

func formatDates(dates []time.Time) string {
  formatted := []string{}
  for _, d := range dates {
    formatted = append(formatted, d.Format("Jan 2"))
  }

  if (len(dates) > 1) {
    commaString := strings.Join(formatted[:len(formatted)-1], ", ")
    return fmt.Sprintf("%s and %s", commaString, formatted[len(formatted)-1])
  } else {
    return formatted[0]
  }
}

func (r *queryResult) String() string {
  year,month,day := r.Date.Date()
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