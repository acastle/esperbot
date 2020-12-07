package events

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/acastle/esperbot/pkg/util"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var ErrInvalidWeekday = errors.New("invalid weekday")

const RecurrentEventIndex string = "index:recurring"

type RecurringEvent struct {
	ID       string
	Name     string
	Weekdays []time.Weekday
}

func RecurringEventKeyForId(id string) string {
	return fmt.Sprintf("recurring:%s", id)
}

func SerializeWeekdays(days []time.Weekday) (string, error) {
	str := make([]string, len(days))
	for i, v := range days {
		str[i] = strconv.Itoa(int(v))
	}

	return strings.Join(str, ","), nil
}

func DeserializeWeekdays(in string) ([]time.Weekday, error) {
	strs := strings.Split(in, ",")
	days := make([]time.Weekday, len(strs))
	for i, s := range strs {
		day, err := strconv.Atoi(s)
		if err != nil || i > 6 {
			return nil, ErrInvalidWeekday
		}

		days[i] = time.Weekday(day)
	}

	return days, nil
}

func UpsertRecurringEvent(redis *redis.Client, event RecurringEvent) error {
	pipe := redis.Pipeline()
	days, err := SerializeWeekdays(event.Weekdays)
	if err != nil {
		return fmt.Errorf("serialize weekdays: %w", err)
	}

	key := RecurringEventKeyForId(event.ID)
	pipe.HSet(key, "id", event.ID)
	pipe.HSet(key, "name", event.Name)
	pipe.HSet(key, "weekdays", days)
	pipe.SAdd(RecurrentEventIndex, event.ID)
	_, err = pipe.Exec()
	if err != nil {
		return fmt.Errorf("execute pipeline: %w", err)
	}

	return nil
}

func GetRecurringEventById(redis *redis.Client, id string) (RecurringEvent, error) {
	key := RecurringEventKeyForId(id)
	result := redis.HGetAll(key)
	if result.Err() != nil {
		return RecurringEvent{}, fmt.Errorf("check for existing recurring event: %w", result.Err())
	}

	data := result.Val()
	days, err := DeserializeWeekdays(data["weekdays"])
	if err != nil {
		return RecurringEvent{}, fmt.Errorf("deserialize weekdays: %w", err)
	}

	return RecurringEvent{
		ID:       data["id"],
		Name:     data["name"],
		Weekdays: days,
	}, nil
}

func GetRecurringEvents(redis *redis.Client) ([]RecurringEvent, error) {
	result := redis.SMembers(RecurrentEventIndex)
	if result.Err() != nil {
		return nil, fmt.Errorf("get recurring events from index: %w", result.Err())
	}

	evtIds := result.Val()
	evts := make([]RecurringEvent, len(evtIds))
	for i, id := range evtIds {
		evt, err := GetRecurringEventById(redis, id)
		if err != nil {
			return nil, fmt.Errorf("get event by id: %w", err)
		}

		evts[i] = evt
	}

	return evts, nil
}

func WeeklyEventsForRecurringEvent(template RecurringEvent, date time.Time) ([]Event, error) {
	evts := []Event{}
	for _, weekday := range template.Weekdays {
		evts = append(evts, Event{
			ID:               uuid.New().String(),
			Name:             template.Name,
			Time:             util.BeginningOfDay(util.DayOfWeek(date.UTC(), weekday)),
			Status:           Unscheduled,
			RecurringEventID: template.ID,
		})
	}

	return evts, nil
}

var ErrDateRangeTooLarge = errors.New("date range must be less than one year")

func ScheduleEventsForWeek(redis *redis.Client, date time.Time) error {
	log.WithFields(log.Fields{
		"begin": util.BeginningOfWeek(date.UTC()),
	}).Info("scheduling events for the week")
	templates, err := GetRecurringEvents(redis)
	if err != nil {
		return fmt.Errorf("get recurring events: %w", err)
	}

	for _, template := range templates {
		evts, err := WeeklyEventsForRecurringEvent(template, date)
		if err != nil {
			return fmt.Errorf("get weekly events: %w", err)
		}

		existingEvents, err := GetEventsForWeek(redis, date)
		if err != nil {
			return fmt.Errorf("get existing events: %w", err)
		}
		for _, evt := range evts {
			if ContainsEventForRecurringEvent(existingEvents, template.ID, evt.Time) {
				continue
			}

			log.WithFields(log.Fields{
				"begin": evt.Time,
				"id":    evt.ID,
			}).Info("scheduling event")
			err = ScheduleEvent(redis, evt)
			if err != nil {
				return fmt.Errorf("schedule event: %w", err)
			}
		}

	}

	return nil
}

func ContainsEventForRecurringEvent(evts []Event, id string, date time.Time) bool {
	for _, evt := range evts {
		if evt.RecurringEventID == id && evt.Time.Equal(date) {
			return true
		}
	}

	return false
}
