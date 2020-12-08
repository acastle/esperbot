package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/acastle/esperbot/pkg/commands"
	"github.com/acastle/esperbot/pkg/util"
)

const CommandPrefix = "!"

var ErrMissingPrefix = errors.New("commands must start with the prefix '!'")
var ErrUnknownCommand = errors.New("unknown command")

func Parse(command string) (commands.Command, error) {
	if !strings.HasPrefix(command, CommandPrefix) {
		return nil, ErrMissingPrefix
	}

	fields := strings.Fields(command)
	switch fields[0] {
	case "!help":
		return &commands.HelpCommand{}, nil
	case "!setname":
		return &commands.SetNameCommand{
			Name: strings.Join(fields[1:], "-"),
		}, nil
	case "!out":
		dates, err := util.FlagsToDateRange(fields[1:])
		if err != nil {
			return nil, fmt.Errorf("parse flags: %w", err)
		}
		return &commands.OutCommand{
			Dates: dates,
		}, nil
	case "!late":
		dates, err := util.FlagsToDateRange(fields[1:])
		if err != nil {
			return nil, fmt.Errorf("parse flags: %w", err)
		}
		return &commands.LateCommand{
			Dates: dates,
		}, nil
	case "!ontime":
		dates, err := util.FlagsToDateRange(fields[1:])
		if err != nil {
			return nil, fmt.Errorf("parse flags: %w", err)
		}
		return &commands.OnTimeCommand{
			Dates: dates,
		}, nil
	case "!in":
		dates, err := util.FlagsToDateRange(fields[1:])
		if err != nil {
			return nil, fmt.Errorf("parse flags: %w", err)
		}
		return &commands.InCommand{
			Dates: dates,
		}, nil
	case "!schedule":
		return &commands.ScheduleCommand{}, nil
	case "!events":
		return &commands.EventsCommand{}, nil
	case "!announce":
		return &commands.AnnounceCommand{}, nil
	default:
		return nil, ErrUnknownCommand
	}

}
