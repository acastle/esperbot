package events

import (
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
)

func TestUpsertRecurringEvent(t *testing.T) {
	svc, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer svc.Close()

	client := redis.NewClient(&redis.Options{
		Addr: svc.Addr(),
	})

	evt := RecurringEvent{
		ID:       "abc123",
		Name:     "foo",
		Weekdays: []time.Weekday{time.Wednesday, time.Thursday},
	}
	err = UpsertRecurringEvent(client, evt)
	isMember, err := svc.SIsMember(RecurrentEventIndex, evt.ID)
	if err != nil {
		t.Error(err)
	}

	if !isMember {
		t.Error("Did not add event to index")
	}

	key := RecurringEventKeyForId(evt.ID)
	if evt.ID != svc.HGet(key, "id") {
		t.Error("did not set id")
	}

	if evt.Name != svc.HGet(key, "name") {
		t.Error("did not set name")
	}

	if "3,4" != svc.HGet(key, "weekdays") {
		t.Error("did not set weekday")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestGetRecurringEventById(t *testing.T) {
	svc, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer svc.Close()

	id := "abc123"
	expected := RecurringEvent{
		ID:       "abc123",
		Name:     "foo",
		Weekdays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday},
	}

	svc.SAdd(RecurrentEventIndex, expected.ID)
	key := RecurringEventKeyForId(id)
	svc.HSet(key, "id", expected.ID)
	svc.HSet(key, "name", expected.Name)
	svc.HSet(key, "weekdays", "1,2,3")

	client := redis.NewClient(&redis.Options{
		Addr: svc.Addr(),
	})

	evt, err := GetRecurringEventById(client, expected.ID)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, evt) {
		t.Errorf("expected '%v' got '%v'", expected, evt)
	}
}

func TestGetRecurringEvents(t *testing.T) {
	svc, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer svc.Close()
	client := redis.NewClient(&redis.Options{
		Addr: svc.Addr(),
	})

	expected := []RecurringEvent{{
		ID:       "abc123",
		Name:     "foo",
		Weekdays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday},
	}}

	svc.SetAdd(RecurrentEventIndex, expected[0].ID)
	key := RecurringEventKeyForId(expected[0].ID)
	svc.HSet(key, "id", expected[0].ID)
	svc.HSet(key, "name", expected[0].Name)
	svc.HSet(key, "weekdays", "1,2,3")

	evts, err := GetRecurringEvents(client)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, evts) {
		t.Errorf("expected '%v' got '%v'", expected, evts)
	}
}
