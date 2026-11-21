package commands

type BaseControlCommand struct {
	BotID     string
	Payload   map[string]any
	Timestamp string
}

type BaseControlCommandI interface {
	IsBaseControlCommand() *BaseControlCommand
}

func (BaseControlCommand *BaseControlCommand) IsBaseControlCommand() *BaseControlCommand {
	return BaseControlCommand
}
