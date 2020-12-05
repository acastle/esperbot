package bot

import (
	"errors"
	"strings"

	"github.com/acastle/esperbot/pkg/commands"
)

const CommandPrefix = "!"

var ErrMissingPrefix = errors.New("commands must start with the prefix '!'")
var ErrUnknownCommand = errors.New("unknown command")

type Command interface {
	Execute() error
}

func Parse(command string) (Command, error) {
	if !strings.HasPrefix(command, CommandPrefix) {
		return nil, ErrMissingPrefix
	}

	fields := strings.Fields(command)
	switch fields[0] {
	case "!help":
		return &commands.HelpCommand{}, nil
	default:
		return nil, ErrUnknownCommand
	}

}
