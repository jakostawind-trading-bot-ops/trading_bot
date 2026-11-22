package di

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/processors/minsev"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

// BootstrapSettingsProvider передаёт настройки запуска в конструкторы Fx.
func BootstrapSettingsProvider(BootstrapSettings *internal.BootstrapSettings) fx.Option {
	return fx.Supply(BootstrapSettings)
}

// serviceAttributes связывает обе разновидности телеметрии с одним сервисом и окружением.
func serviceAttributes(Stage string) []attribute.KeyValue {
	Attributes := []attribute.KeyValue{attribute.String("service.name", "trading-bot")}
	if Environment := strings.TrimSpace(Stage); Environment != "" {
		Attributes = append(Attributes, attribute.String("deployment.environment.name", Environment))
	}
	return Attributes
}

// NewLogger создаёт контекстный slog-логгер и возвращает SDK для выгрузки при остановке.
// Settings задаёт порог DEBUG или INFO и окружение. Output получает JSON-записи OTel.
//
//	slog -> otelslog -> minsev -> batch -> stdoutlog
func NewLogger(Settings *internal.BootstrapSettings, Output io.Writer) (*slog.Logger, *sdklog.LoggerProvider, error) {
	var Level slog.Level
	switch Settings.LoggerLevel {
	case "DEBUG":
		Level = slog.LevelDebug
	case "INFO":
		Level = slog.LevelInfo
	default:
		return nil, nil, fmt.Errorf("недопустимый LOGGER_LEVEL %q, ожидается DEBUG или INFO", Settings.LoggerLevel)
	}
	Exporter, ExportError := stdoutlog.New(stdoutlog.WithWriter(Output))
	if ExportError != nil {
		return nil, nil, ExportError
	}
	Provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(resource.NewSchemaless(serviceAttributes(Settings.Stage)...)),
		sdklog.WithProcessor(minsev.NewLogProcessor(sdklog.NewBatchProcessor(Exporter), minsev.Severity(Level))),
	)
	return otelslog.NewLogger("trading_bot", otelslog.WithLoggerProvider(Provider), otelslog.WithSource(true)), Provider, nil
}

// LoggerProvider создаёт общий логгер при сборке Fx и выгружает накопленные записи после остановки потребителей.
func LoggerProvider(Lifecycle fx.Lifecycle, Settings *internal.BootstrapSettings) (*slog.Logger, error) {
	Logger, Provider, SetupError := NewLogger(Settings, os.Stdout)
	if SetupError != nil {
		return nil, SetupError
	}
	Lifecycle.Append(fx.Hook{OnStop: Provider.Shutdown})
	return Logger, nil
}

// RegisterTracing регистрирует запуск SDK до NATS и его остановку после обработчиков.
// Каждому запуску соответствует отдельный service.instance.id. Дочерние спаны
// наследуют решение родителя, все новые корневые трейсы записываются.
func RegisterTracing(Lifecycle fx.Lifecycle, Settings *internal.BootstrapSettings) {
	var Provider *sdktrace.TracerProvider
	Lifecycle.Append(fx.Hook{
		OnStart: func(Context context.Context) error {
			InstanceID, GenerateError := uuid.NewRandom()
			if GenerateError != nil {
				return GenerateError
			}
			Attributes := append(serviceAttributes(Settings.Stage), attribute.String("service.instance.id", InstanceID.String()))
			Resource, ResourceError := resource.New(Context, resource.WithTelemetrySDK(), resource.WithAttributes(Attributes...))
			if ResourceError != nil {
				return ResourceError
			}
			Exporter, ExportError := stdouttrace.New()
			if ExportError != nil {
				return ExportError
			}
			Provider = sdktrace.NewTracerProvider(
				sdktrace.WithResource(Resource),
				sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
				sdktrace.WithBatcher(Exporter),
			)
			otel.SetTracerProvider(Provider)
			return nil
		},
		OnStop: func(Context context.Context) error { return Provider.Shutdown(Context) },
	})
}
