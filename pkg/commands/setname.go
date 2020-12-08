package commands

import (
	"fmt"

	"github.com/acastle/esperbot/pkg/events"
	log "github.com/sirupsen/logrus"
)

type SetNameCommand struct {
	Name string
}

func (c SetNameCommand) Execute(ctx Context) error {
	log.WithField("id", ctx.Sender.ID).Info("set user alias")
	err := events.SetUserAlias(ctx.Redis, ctx.Sender.ID, c.Name)
	if err != nil {
		return fmt.Errorf("set user name: %w", err)
	}

	_, err = ctx.Session.ChannelMessageSend(ctx.ChannelID, fmt.Sprintf("From this day forward we call you '%s'... I hope you are happy.", c.Name))
	if err != nil {
		return fmt.Errorf("send response: %w", err)
	}

	return nil
}
