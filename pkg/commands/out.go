package commands

import "github.com/acastle/esperbot/pkg/util"

type OutCommand struct {
	Dates util.DateRange
}

func (c OutCommand) Execute(ctx Context) error {
	return nil
}
