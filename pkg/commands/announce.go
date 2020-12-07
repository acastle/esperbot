package commands

import (
	"fmt"
	"time"

	"github.com/acastle/esperbot/pkg/events"
	log "github.com/sirupsen/logrus"
)

type AnnounceCommand struct {
}

func (c AnnounceCommand) Execute(ctx Context) error {
	evts, err := events.GetEventsForWeek(ctx.Redis, time.Now().UTC())
	if err != nil {
		log.Error(err)
	}

	for _, evt := range evts {
		evt.AnnounceChannelID = ctx.ChannelID
		err := events.AnnounceEvent(ctx.Session, ctx.Redis, evt)
		if err != nil {
			return fmt.Errorf("announce event: %w", err)
		}
	}

	return nil
}
