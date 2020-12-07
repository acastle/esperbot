package commands

import "fmt"

const HelpText string = `
!help - prints the available commands and their usage.
`

type HelpCommand struct {
}

func (h HelpCommand) Execute(ctx Context) error {
	_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID, HelpText)
	if err != nil {
		return fmt.Errorf("send help message: %w", err)
	}
	return nil
}
