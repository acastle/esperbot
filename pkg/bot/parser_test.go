package bot

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name    string
		command string
		expErr  error
	}{
		{
			"help command",
			"!help",
			nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Parse(c.command)
			if errors.Is(err, c.expErr) {
				t.Error("expected an error")
			}
		})
	}

}
