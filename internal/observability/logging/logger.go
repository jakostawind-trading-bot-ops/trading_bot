package observability_logging

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/processors/minsev"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Logger задаёт параметры экспорта логов одного сервиса.
type Logger struct {
	ServiceName string // Имя сервиса в OTel resource.
	Stage       string // Окружение из STAGE, если оно задано.
	LoggerLevel string // Минимальный уровень из LOGGER_LEVEL.
}

// GetLevel преобразует значение LOGGER_LEVEL в порог фильтра OTel.
func (Logger *Logger) GetLevel() (slog.Level, error) {
	switch Logger.LoggerLevel {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	default:
		return slog.LevelInfo, fmt.Errorf(
			"недопустимый LOGGER_LEVEL %q, ожидается DEBUG или INFO",
			Logger.LoggerLevel,
		)
	}
}

// CreateOTelLogger создаёт slog-логгер с экспортом OTel-записей в Out.
// Out — обязательный writer; DI передаёт os.Stdout.
// Возвращает логгер и SDK-провайдер, чей Shutdown обязан вызвать владелец.
// Ошибка означает некорректный LoggerLevel либо отказ создания exporter.
// Логгер передаёт ctx каждой записи в SDK для извлечения trace_id и span_id.
// Формат JSON задаёт stdoutlog.
//
// Этапы:
//  1. Проверить уровень и создать stdout exporter.
//  2. Описать сервис и создать SDK с фильтром уровней перед batch processor.
//  3. Подключить otelslog к созданному SDK без глобальной регистрации.
//
// Схема:
//
//	slog -> otelslog -> SDK -> minsev -> batch -> stdoutlog -> Out
func (Logger *Logger) CreateOTelLogger(Out io.Writer) (*slog.Logger, *sdklog.LoggerProvider, error) {
	// Этап 1: ошибочные настройки отклоняем до создания фонового processor.
	Level, LevelError := Logger.GetLevel()
	if LevelError != nil {
		return nil, nil, LevelError
	}
	Exporter, ExportError := stdoutlog.New(stdoutlog.WithWriter(Out))
	if ExportError != nil {
		return nil, nil, fmt.Errorf("logging: создание stdout exporter: %w", ExportError)
	}

	// Этап 2: resource связывает записи с сервисом и окружением приложения.
	Attributes := []attribute.KeyValue{
		attribute.String("service.name", Logger.ServiceName),
	}
	if Stage := strings.TrimSpace(Logger.Stage); Stage != "" {
		Attributes = append(Attributes, attribute.String("deployment.environment.name", Stage))
	}
	Provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource.NewSchemaless(Attributes...)),
		sdklog.WithProcessor(minsev.NewLogProcessor(
			sdklog.NewBatchProcessor(Exporter),
			// minsev использует шкалу slog: DEBUG = -4, INFO = 0.
			minsev.Severity(Level),
		)),
	)

	// Этап 3: прикладной код продолжает получать обычный *slog.Logger.
	return otelslog.NewLogger(
		"trading_bot",
		otelslog.WithLoggerProvider(Provider),
		otelslog.WithSource(true),
	), Provider, nil
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
