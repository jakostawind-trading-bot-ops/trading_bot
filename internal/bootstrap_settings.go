package internal

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// BootstrapSettings содержит настройки запуска из переменных окружения.
type BootstrapSettings struct {
	Stage        string // Окружение приложения из STAGE.
	LoggerLevel  string // Минимальный уровень OTel Logs из LOGGER_LEVEL (DEBUG или INFO).
	NatsUrl      string // Адрес NATS из NATS_URL.
	NatsUser     string // Имя пользователя NATS из NATS_USER.
	NatsPassword string // Пароль NATS из NATS_PASSWORD; не предназначен для логов.
}

// LoadBootstrapSettings возвращает настройки из окружения без проверки значений.
// Отсутствующие строковые настройки остаются пустыми; секреты не логируются.
func LoadBootstrapSettings() BootstrapSettings {
	return BootstrapSettings{
		Stage:        os.Getenv("STAGE"),
		LoggerLevel:  os.Getenv("LOGGER_LEVEL"),
		NatsUrl:      os.Getenv("NATS_URL"),
		NatsUser:     os.Getenv("NATS_USER"),
		NatsPassword: os.Getenv("NATS_PASSWORD"),
	}
}

// FinishSpan записывает ошибку операции и завершает Span. Указатель на именованный
// результат позволяет вызвать defer до выполнения операции и учесть все ранние возвраты.
func FinishSpan(Span trace.Span, OperationError *error) {
	if *OperationError != nil {
		Span.RecordError(*OperationError)
		Span.SetStatus(codes.Error, (*OperationError).Error())
	}
	Span.End()
}

// contractError сохраняет исходную ошибку JSON для errors.Is/As и безопасный текст для телеметрии.
type contractError struct {
	Description string // Описание операции без содержимого payload.
	Cause       error  // Исходная ошибка кодека.
}

// Error возвращает безопасное описание для спанов и обёрток ошибок.
func (Failure contractError) Error() string { return Failure.Description }

// Unwrap сохраняет доступ к исходной ошибке кодека.
func (Failure contractError) Unwrap() error { return Failure.Cause }

// LogValue добавляет категорию ошибки JSON и положение в сообщении, не раскрывая значения полей.
func (Failure contractError) LogValue() slog.Value {
	var SyntaxError *json.SyntaxError
	var TypeError *json.UnmarshalTypeError
	switch {
	case errors.As(Failure.Cause, &SyntaxError):
		return slog.GroupValue(slog.String("kind", "json_syntax"), slog.Int64("byte_offset", SyntaxError.Offset))
	case errors.As(Failure.Cause, &TypeError):
		return slog.GroupValue(slog.String("kind", "json_type"), slog.Int64("byte_offset", TypeError.Offset))
	default:
		return slog.StringValue(Failure.Description)
	}
}

// SafeError оборачивает Cause, чтобы произвольные данные JSON не оказались в логе или спане.
func SafeError(Description string, Cause error) error {
	return contractError{Description: Description, Cause: Cause}
}
