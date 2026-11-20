package lifecycle

import (
	"fmt"
	"strings"
)

type Lifecycle struct {
	GeneralState BotStatus
}

type BotStatus interface {
	GetBotStatus() LifecycleStatus
	SetBotStatus(LifecycleStatus)
}

func (Lifecycle *Lifecycle) EnsureStatus(Allowed ...LifecycleStatus) error {
	CurrentStatus := Lifecycle.GeneralState.GetBotStatus()

	for _, Status := range Allowed {
		if CurrentStatus == Status {
			return nil
		}
	}

	values := make([]string, 0, len(Allowed))
	for _, status := range Allowed {
		values = append(values, string(status))
	}

	return fmt.Errorf(
		"cannot change status from %s; allowed statuses: %s",
		CurrentStatus,
		strings.Join(values, ", "),
	)
}

func (Lifecycle *Lifecycle) ChangeStatus(
	NewStatus LifecycleStatus,
	Allowed ...LifecycleStatus,
) {
	Lifecycle.EnsureStatus(Allowed...)

	// PreviousStatus := Lifecycle.GeneralState.GetBotStatus()
	Lifecycle.GeneralState.SetBotStatus(NewStatus)

	// logger.info(
    //         "Статус lifecycle бота изменён "
    //         "bot_id=%s previous_status=%s new_status=%s",
    //         self.general_state.bot_id,
    //         previous_status.value,
    //         new_status.value,
    //     )
}
