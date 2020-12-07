package commands

import (
	"time"

	"github.com/acastle/esperbot/pkg/events"
	log "github.com/sirupsen/logrus"
)

type ScheduleCommand struct {
}

func (c ScheduleCommand) Execute(ctx Context) error {
	err := events.ScheduleEventsForWeek(ctx.Redis, time.Now())
	if err != nil {
		log.Error(err)
	}
	return nil
}
