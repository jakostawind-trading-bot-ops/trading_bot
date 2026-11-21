package commands

import "context"

type ControlCommandHandlerI interface {
	HandleControlCommand(
		ctx context.Context,
		ControlCommand BaseControlCommandI,
	) error
}
