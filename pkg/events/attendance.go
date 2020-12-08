package events

import (
	"fmt"
	"time"

	"github.com/acastle/esperbot/pkg/util"
	"github.com/go-redis/redis"
)

type Attendance struct {
	Absent []string
	Late   []string
}

type UserListType string

var Absent UserListType = "absent"
var Late UserListType = "late"

func UserListKeyForDate(date time.Time, t UserListType) string {
	return fmt.Sprintf("%s:%d", t, util.BeginningOfDay(date.UTC()).Unix())
}

func UserListKeyForEventId(id string, t UserListType) string {
	return fmt.Sprintf("event:%s:%s", id, t)
}

func GetAttendanceForDay(redis *redis.Client, date time.Time) (Attendance, error) {
	result := redis.SMembers(UserListKeyForDate(date, Absent))
	if result.Err() != nil {
		return Attendance{}, fmt.Errorf("lookup absent: %w", result.Err())
	}

	absent := result.Val()
	result = redis.SMembers(UserListKeyForDate(date, Late))
	if result.Err() != nil {
		return Attendance{}, fmt.Errorf("lookup late: %w", result.Err())
	}

	late := result.Val()
	return Attendance{
		Absent: absent,
		Late:   late,
	}, nil
}

func GetAttendanceForEvent(redis *redis.Client, evt Event) (Attendance, error) {
	result := redis.SMembers(UserListKeyForEventId(evt.ID, Absent))
	if result.Err() != nil {
		return Attendance{}, fmt.Errorf("lookup absent: %w", result.Err())
	}

	absent := result.Val()
	result = redis.SMembers(UserListKeyForEventId(evt.ID, Late))
	if result.Err() != nil {
		return Attendance{}, fmt.Errorf("lookup late: %w", result.Err())
	}

	late := result.Val()
	return Attendance{
		Absent: absent,
		Late:   late,
	}, nil
}

func UserListAddForRange(redis *redis.Client, r util.DateRange, id string, t UserListType) error {
	return util.ForEachDay(r, func(d time.Time) error {
		err := UserListAdd(redis, d, id, t)
		if err != nil {
			return fmt.Errorf("add to user list: %w", err)
		}

		return nil
	})
}

func UserListRemoveForRange(redis *redis.Client, r util.DateRange, id string, t UserListType) error {
	return util.ForEachDay(r, func(d time.Time) error {
		err := UserListRemove(redis, d, id, t)
		if err != nil {
			return fmt.Errorf("remove from user list: %w", err)
		}

		return nil
	})
}

func UserListAdd(redis *redis.Client, date time.Time, id string, t UserListType) error {
	key := UserListKeyForDate(date, t)
	result := redis.SAdd(key, id)
	if result.Err() != nil {
		return fmt.Errorf("add id to set: %w", result.Err())
	}

	evts, err := GetEventsForWeek(redis, date)
	if err != nil {
		return fmt.Errorf("get events for week: %w", err)
	}

	for _, evt := range evts {
		err := EventUserListAdd(redis, evt, id, t)
		if err != nil {
			return fmt.Errorf("mark absent for event: %w", err)
		}
	}

	return nil
}

func UserListRemove(redis *redis.Client, date time.Time, id string, t UserListType) error {
	key := UserListKeyForDate(date, t)
	result := redis.SAdd(key, id)
	if result.Err() != nil {
		return fmt.Errorf("add id to set: %w", result.Err())
	}

	evts, err := GetEventsForWeek(redis, date)
	if err != nil {
		return fmt.Errorf("get events for week: %w", err)
	}

	for _, evt := range evts {
		err := EventUserListAdd(redis, evt, id, t)
		if err != nil {
			return fmt.Errorf("mark absent for event: %w", err)
		}
	}

	return nil
}

func EventUserListAdd(redis *redis.Client, evt Event, id string, t UserListType) error {
	key := UserListKeyForEventId(evt.ID, t)
	result := redis.SAdd(key, id)
	if result.Err() != nil {
		return fmt.Errorf("add id to set: %w", result.Err())
	}

	return nil
}

func EventUserListRemove(redis *redis.Client, evt Event, id string, t UserListType) error {
	key := UserListKeyForEventId(evt.ID, t)
	result := redis.SRem(key, id)
	if result.Err() != nil {
		return fmt.Errorf("remove id from set: %w", result.Err())
	}

	return nil
}
