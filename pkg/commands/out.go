package commands

import (
	"fmt"

	"github.com/acastle/esperbot/pkg/events"
	"github.com/acastle/esperbot/pkg/util"
	log "github.com/sirupsen/logrus"
)

type OutCommand struct {
	Dates util.DateRange
}

func (c OutCommand) Execute(ctx Context) error {
	err := events.UserListAddForRange(ctx.Redis, c.Dates, ctx.Sender.ID, events.Absent)
	if err != nil {
		return fmt.Errorf("mark user absent for day: %w", err)
	}

	evts, err := events.GetEventsForDateRange(ctx.Redis, c.Dates)
	if err != nil {
		return fmt.Errorf("get events for range: %w", err)
	}

	log.WithFields(log.Fields{
		"begin": c.Dates.Begin,
		"end":   c.Dates.End,
	}).Info("mark user out for range")
	for _, evt := range evts {
		log.WithFields(log.Fields{
			"user":  ctx.Sender.ID,
			"list":  events.Absent,
			"event": evt.ID,
		}).Info("add user to event list")
		err = events.EventUserListAdd(ctx.Redis, evt, ctx.Sender.ID, events.Absent)
		if err != nil {
			return fmt.Errorf("add user to user list: %w", err)
		}

		err = events.AnnounceEvent(ctx.Session, ctx.Redis, evt)
		if err != nil {
			return fmt.Errorf("announce event: %w", err)
		}
	}

	alias, err := events.GetUserAlias(ctx.Redis, ctx.Session, ctx.Sender.ID)
	if err != nil {
		return fmt.Errorf("fetch user alias: %w", err)
	}

	_, err = ctx.Session.ChannelMessageSend(ctx.ChannelID, fmt.Sprintf("Marked '%s' out for all events between %s and %s", alias, c.Dates.Begin.Format(StandardDateFormat), c.Dates.End.Format(StandardDateFormat)))
	if err != nil {
		return fmt.Errorf("send response: %w", err)
	}

	return nil
}
