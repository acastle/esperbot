package main

import (
  "time"
  "errors"
  "fmt"
  "strings"
  "github.com/acastle/apiai-go"
  "github.com/bwmarrin/discordgo"
)

var (
  ErrInvalidRaidDay = errors.New("Not a valid raid day")
  ErrNoDatesSpecified = errors.New("Need to specifiy dates")
)

func IsValidRaidDay(weekday time.Weekday) bool {
  return weekday != 2 || weekday != 3
}

func getSetKey(time time.Time) string {
  year,month,day := time.Date()
  return fmt.Sprintf("attendance:%d:%d:%d", year, month, day)
}

func getDates(resp apiaigo.ResponseStruct) ([]time.Time, error){
  dateStrings := resp.Result.Parameters["date"].Values
  if dateStrings == nil {
    return []time.Time{}, ErrNoDatesSpecified
  }

  var dates []time.Time
  for _,s := range dateStrings {
    date, err := time.Parse("2006-01-02", s)
    if err != nil {
      return []time.Time{}, err
    }

    dates = append(dates, date)
  }

  return dates, nil
}

type listParam struct {
  data []interface{}
}

type MissRaidResult struct {
  Name string
  Dates []time.Time
}

func MissRaid(user *discordgo.User, resp apiaigo.ResponseStruct) ([]MissRaidResult, error) {
  members := resp.Result.Parameters["member"].Values
  if len(members) == 0 {
    members = []string{user.Username}
  }

  dates, err := getDates(resp)
  if err != nil {
    return []MissRaidResult{}, err
  }

  results := []MissRaidResult{}
  for _,n := range members {
    val := strings.Title(n)

    for _,d := range dates {
      if !IsValidRaidDay(d.Weekday()) {
        return []MissRaidResult{},ErrInvalidRaidDay
      }

      key := getSetKey(d)
      result := Redis.SAdd(key, val)
      if result.Err() != nil {
        return []MissRaidResult{},result.Err()
      }
    }

    results = append(results, MissRaidResult{val, dates})
  }

  return results, nil
}

type QueryResult struct {
  Date time.Time
  Members []string
}

func Query(resp apiaigo.ResponseStruct) ([]QueryResult, error){
  results := []QueryResult{}
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
    results = append(results, QueryResult{d, result.Val()})
  }

  return results, nil
}