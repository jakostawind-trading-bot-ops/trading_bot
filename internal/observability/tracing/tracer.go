package observability_tracing

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Tracing управляет SDK OpenTelemetry и экспортом spans в stdout для разработки.
// Создаётся в DI; Start и Shutdown вызываются последовательно через lifecycle.
// Настройки задаются до Start. После успешного запуска повторный Start запрещён,
// в том числе после Shutdown. Один экземпляр обслуживает всё приложение.
type Tracing struct {
	ServiceName    string                   // Имя сервиса в resource; обязательно, например trading-bot.
	Stage          string                   // Окружение приложения; пустое значение не добавляется в resource.
	TracerProvider *sdktrace.TracerProvider // SDK-провайдер; nil до успешного Start, сохраняется после Shutdown.
}

// Start создаёт SDK-провайдер и делает его доступным через otel.Tracer.
//
// Context — ненулевой контекст инициализации: отмена проверяется при входе,
// затем контекст передаётся resource.New. Он не управляет временем жизни SDK.
// Возвращает ошибку повторного запуска, отмены Context, пустого ServiceName
// либо создания resource/exporter; nil означает завершённую настройку.
//
// Побочные эффекты: запускает batch processor и заменяет глобальный provider.
// Sampling включает запись всех новых корневых трейсов; дочерние наследуют
// решение родителя. Это настройка для разработки.
//
// Этапы:
//  1. Проверить состояние, контекст и настройки.
//  2. Создать описание сервиса и stdout exporter.
//  3. Создать SDK с пакетной отправкой spans и зарегистрировать его глобально.
//
// Схема:
//
//	[Проверки] -> [Resource] -> [Exporter] -> [SDK + batch processor]
//	     |            |            |                  |
//	   ошибка       ошибка       ошибка               v
//	     +------------+------------+-> error   [otel.SetTracerProvider]
func (Tracing *Tracing) Start(Context context.Context) error {
	// Этап 1: повторный запуск не должен создавать ещё один batch processor.
	if Tracing.TracerProvider != nil {
		return fmt.Errorf("tracing уже был запущен")
	}
	if ContextError := Context.Err(); ContextError != nil {
		return ContextError
	}
	ServiceName := strings.TrimSpace(Tracing.ServiceName)
	if ServiceName == "" {
		return fmt.Errorf("tracing: имя сервиса не задано")
	}

	// Этап 2: resource описывает приложение и одинаков для всех его spans.
	InstanceID, InstanceError := uuid.NewRandom()
	if InstanceError != nil {
		return fmt.Errorf("tracing: создание service.instance.id: %w", InstanceError)
	}

	Attributes := []attribute.KeyValue{
		attribute.String("service.name", ServiceName),
		attribute.String("service.instance.id", InstanceID.String()),
	}

	if Stage := strings.TrimSpace(Tracing.Stage); Stage != "" {
		Attributes = append(Attributes,
			attribute.String("deployment.environment.name", Stage),
		)
	}

	Resource, ResourceError := resource.New(Context,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(Attributes...),
	)
	if ResourceError != nil {
		return fmt.Errorf("tracing: создание resource: %w", ResourceError)
	}
	Exporter, ExportError := stdouttrace.New()
	if ExportError != nil {
		return fmt.Errorf("tracing: создание stdout exporter: %w", ExportError)
	}

	// Этап 3: SDK владеет processor и exporter; освобождаем их через Shutdown.
	Provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(Resource),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
		sdktrace.WithBatcher(Exporter),
	)
	Tracing.TracerProvider = Provider
	otel.SetTracerProvider(Provider)
	return nil
}

// Shutdown завершает экспорт spans и освобождает ресурсы SDK.
//
// Context — ненулевой контекст остановки, передаваемый SDK для ограничения ожидания.
// Возвращает ошибку SDK либо nil, если SDK остановлен или ещё не был создан.
// Вызывается после завершения обработчиков событий. При таймауте доставка
// оставшихся spans не гарантируется. Повторный вызов не повторяет их отправку.
func (Tracing *Tracing) Shutdown(Context context.Context) error {
	if Tracing.TracerProvider == nil {
		return nil
	}
	return Tracing.TracerProvider.Shutdown(Context)
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
