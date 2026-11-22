package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"go.opentelemetry.io/otel"
)

// Lifecycle проверяет и меняет статус общего состояния бота.
type Lifecycle struct {
	GeneralState BotStatus    // Состояние экземпляра в памяти процесса.
	Logger       *slog.Logger // Логгер жизненного цикла с идентификатором бота.
}

// BotStatus предоставляет доступ к текущему статусу экземпляра.
type BotStatus interface {
	GetBotStatus() LifecycleStatus
	SetBotStatus(LifecycleStatus)
}

// EnsureStatus сверяет текущий статус со списком Allowed в контексте операции.
// Возвращает ошибку с текущим и допустимыми статусами при несовпадении.
func (Lifecycle *Lifecycle) EnsureStatus(Context context.Context, Allowed ...LifecycleStatus) (StatusError error) {
	Context, Span := otel.Tracer("trading_bot/internal/lifecycle").Start(Context, "Lifecycle.EnsureStatus")
	defer tracing.FinishSpan(Span, &StatusError)
	CurrentStatus := Lifecycle.GeneralState.GetBotStatus()
	Lifecycle.Logger.DebugContext(Context, "Проверка текущего статуса", "operation", "ensure_status", "bot_status", CurrentStatus)
	for _, Status := range Allowed {
		if CurrentStatus == Status {
			return nil
		}
	}
	AllowedNames := make([]string, 0, len(Allowed))
	for _, Status := range Allowed {
		AllowedNames = append(AllowedNames, string(Status))
	}
	return fmt.Errorf("cannot change status from %s; allowed statuses: %s", CurrentStatus, strings.Join(AllowedNames, ", "))
}

// ChangeStatus устанавливает NewStatus, если текущий статус входит в Allowed.
// При недопустимом переходе возвращает ошибку и сохраняет текущий статус.
func (Lifecycle *Lifecycle) ChangeStatus(Context context.Context, NewStatus LifecycleStatus, Allowed ...LifecycleStatus) (StatusError error) {
	Context, Span := otel.Tracer("trading_bot/internal/lifecycle").Start(Context, "Lifecycle.ChangeStatus")
	defer tracing.FinishSpan(Span, &StatusError)

	if StatusError = Lifecycle.EnsureStatus(Context, Allowed...); StatusError != nil {
		return StatusError
	}

	PreviousStatus := Lifecycle.GeneralState.GetBotStatus()
	Lifecycle.GeneralState.SetBotStatus(NewStatus)
	Lifecycle.Logger.InfoContext(Context, "Статус бота изменён",
		"operation", "change_status",
		"previous_status", PreviousStatus,
		"new_status", NewStatus,
	)
	return nil
}