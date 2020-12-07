package parser

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/acastle/esperbot/pkg/commands"
	"github.com/acastle/esperbot/pkg/util"
)

func TestParse(t *testing.T) {
	now := time.Now().UTC()
	cases := []struct {
		name       string
		command    string
		expErr     error
		expCommand commands.Command
	}{
		/*{
			"no prefix",
			"help",
			ErrMissingPrefix,
			nil,
		},
		{
			"help command",
			"!help",
			nil,
			&commands.HelpCommand{},
		},*/
		{
			"out command no args",
			"!out",
			nil,
			&commands.OutCommand{
				Dates: util.DateRange{
					Begin: util.BeginningOfWeek(now),
					End:   util.EndOfWeek(now),
				},
			},
		},
		{
			"in command",
			"!in",
			ErrUnknownCommand,
			nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd, err := Parse(c.command)
			if !errors.Is(err, c.expErr) {
				t.Error("did not return the expected error")
				return
			}

			if !reflect.DeepEqual(cmd, c.expCommand) {
				t.Error("did not correctly parse command")
			}
		})
	}

}
