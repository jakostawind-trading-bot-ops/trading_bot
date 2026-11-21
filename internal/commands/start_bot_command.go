package commands

// StartBotCommand запрашивает выполнение существующего сценария запуска бота.
type StartBotCommand struct {
	BaseControlCommand // Адресат, данные и время команды Control Plane.
}
