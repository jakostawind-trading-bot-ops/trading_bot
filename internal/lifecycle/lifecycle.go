package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	telemetry "github.com/jakostawind-trading-bot-ops/trading_bot/internal"
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
	defer telemetry.FinishSpan(Span, &StatusError)
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

// ChangeStatus устанавливает NewStatus и записывает предыдущий статус.
// Как и прежде, несовпадение с Allowed не блокирует переход. Context связывает проверку и изменение.
func (Lifecycle *Lifecycle) ChangeStatus(Context context.Context, NewStatus LifecycleStatus, Allowed ...LifecycleStatus) {
	Context, Span := otel.Tracer("trading_bot/internal/lifecycle").Start(Context, "Lifecycle.ChangeStatus")
	defer Span.End()
	Logger := Lifecycle.Logger.With("operation", "change_status", "previous_status", Lifecycle.GeneralState.GetBotStatus(), "new_status", NewStatus)
	if StatusError := Lifecycle.EnsureStatus(Context, Allowed...); StatusError != nil {
		Logger.WarnContext(Context, "Переход из статуса вне списка допустимых", "error", StatusError)
	}
	Lifecycle.GeneralState.SetBotStatus(NewStatus)
	Logger.InfoContext(Context, "Статус бота изменён")
}
