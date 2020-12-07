package commands

import (
	"fmt"
	"time"

	"github.com/acastle/esperbot/pkg/events"
	"github.com/dustin/go-humanize"
	log "github.com/sirupsen/logrus"
)

type EventsCommand struct {
}

func (c EventsCommand) Execute(ctx Context) error {
	evts, err := events.GetEventsForWeek(ctx.Redis, time.Now().UTC())
	if err != nil {
		log.Error(err)
	}

	content := "__**Upcoming events:**__\n"
	for _, evt := range evts {
		content = content + fmt.Sprintf("%s(%s) %s", evt.Name, evt.ID, humanize.Time(evt.Time.Add(time.Hour*5)))
	}

	ctx.Session.ChannelMessageSend(ctx.ChannelID, content)
	return nil
}
