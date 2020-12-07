package events

import (
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
)

func TestGetAttendanceForDay(t *testing.T) {
	svc, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer svc.Close()
	client := redis.NewClient(&redis.Options{
		Addr: svc.Addr(),
	})

	day := time.Date(2010, 12, 20, 12, 30, 0, 0, time.UTC)
	svc.SAdd(UserListKeyForDate(day, Absent), "abc123")
	svc.SAdd(UserListKeyForDate(day, Late), "321asd")
	expected := Attendance{
		Absent: []string{"abc123"},
		Late:   []string{"321asd"},
	}

	result, err := GetAttendanceForDay(client, day)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected '%v' got '%v'", expected, result)
	}
}
