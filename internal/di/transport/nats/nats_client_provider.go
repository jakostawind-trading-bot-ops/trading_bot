package di_transport_nats

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/fx"
)

// NatsClient управляет фоновым клиентом NATS в жизненном цикле fx.
// Создаётся через NatsClientProvider; run владеет соединением и закрывает его
// при остановке. Наличие указателя natsClient не гарантирует доступность NATS:
// клиент может находиться в процессе переподключения.
// После запуска BootstrapSettings нельзя менять конкурентно с run.
// Экземпляр не следует копировать после начала использования atomic.Pointer.
// Порядок запуска и остановки задаётся lifecycle fx.
//
// Схема:
//
//	fx.OnStart -> [run в goroutine] -> [клиент NATS]
//	fx.OnStop  -> cancel -> [завершение run] -> закрытие done
type NatsClient struct {
	Logger            *slog.Logger               // Логгер соединения, адрес и учётные данные в него не передаются.
	BootstrapSettings internal.BootstrapSettings // Копия настроек; run использует NatsUrl, NatsUser и NatsPassword.
	Connection        atomic.Pointer[nats.Conn]  // Текущий клиент; nil до создания и после обработки его закрытия.
	Cancel            context.CancelFunc         // Отмена run; устанавливается в OnStart, до запуска равна nil.
	Done              chan struct{}              // Сигнал выхода из run; создаётся в OnStart и закрывается только run.
}

// NatsClientProvider создаёт владельца клиента NATS и регистрирует хуки fx.
// Сам вызов провайдера не подключается к NATS. OnStart запускает run в фоне
// и возвращает nil, не дожидаясь соединения и не передавая ошибки подключения fx.
//
// Параметры:
//   - Lifecycle — lifecycle fx, в который добавляются OnStart и OnStop; должен быть задан.
//   - Logger — общий логгер, используемый для операций соединения.
//   - BootstrapSettings — настройки подключения, копируемые в NatsClient;
//     nil недопустим и приведёт к panic при разыменовании.
//
// Результаты:
//   - *NatsClient — экземпляр с зарегистрированными хуками, ещё без соединения.
//
// Побочные эффекты:
// OnStart создаёт отдельный context от context.Background: контекст запуска fx
// не управляет временем жизни run. OnStop отменяет этот context и ждёт done.
// Если запуск не состоялся, OnStop возвращает nil; если его ctx завершился раньше
// ожидания, возвращает Context.Err(), не гарантируя к этому моменту завершение run.
//
// Этапы:
//  1. Скопировать настройки в новый экземпляр.
//  2. Зарегистрировать запуск и остановку фоновой goroutine.
//
// Схема:
//
//	[Копия настроек] -> [Регистрация хуков] -> NatsClient
//	OnStart -> [context, cancel, done] -> [go run] -> nil
//	OnStop -> [cancel задан?]
//	             +-- нет -> nil
//	             +-- да -> cancel -> [ожидание]
//	                                    +-- done -> nil
//	                                    +-- ctx.Done -> Context.Err()
func NatsClientProvider(
	Lifecycle fx.Lifecycle,
	Logger *slog.Logger,
	BootstrapSettings *internal.BootstrapSettings,
) *NatsClient {
	// Этап 1: сохраняем собственную копию значений настроек для фонового клиента.
	NatsClient := &NatsClient{
		BootstrapSettings: *BootstrapSettings,
		Logger:            Logger.With("component", "nats_client"),
	}

	// Этап 2: передаём управление временем жизни клиента lifecycle fx.
	Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			Context, Cancel := context.WithCancel(context.Background())
			NatsClient.Cancel = Cancel
			NatsClient.Done = make(chan struct{})
			go NatsClient.Run(Context)
			return nil
		},
		OnStop: func(Context context.Context) error {
			if NatsClient.Cancel == nil {
				return nil
			}
			NatsClient.Cancel()
			select {
			case <-NatsClient.Done:
				return nil
			case <-Context.Done():
				return Context.Err()
			}
		},
	})

	return NatsClient
}

// Run создаёт клиент NATS, наблюдает за его закрытием и повторяет создание.
// Вызывается один раз из OnStart после инициализации NatsClient.Done; при любом выходе
// закрывает NatsClient.Done, позволяя OnStop завершить ожидание.
//
// Параметры:
//   - Context — контекст времени жизни клиента; должен быть ненулевым. Отмена
//     прерывает ожидание закрытия и паузу между попытками. В nats.Connect
//     контекст не передаётся, поэтому текущий вызов Connect напрямую не отменяется.
//
// Побочные эффекты:
// Создаёт сетевой клиент, обновляет NatsClient.Connection и пишет структурированные события через slog.
// Ошибки создания, включая некорректный URL, логируются и повторяются до отмены.
// NATS выполняет внутренние переподключения без ограничения числа попыток;
// внешний цикл создаёт новый клиент только после ошибки Connect или ClosedHandler.
// При отмене созданное соединение закрывается через Close, без Drain.
//
// Этапы:
//  1. Настроить подключение, внутренние повторы NATS и обработчики событий.
//  2. Обработать ошибку либо сохранить клиент и ждать отмены или его закрытия.
//  3. Перед новым созданием выдержать отменяемую паузу в две секунды.
//
// Схема:
//
//	[ctx активен?] -- нет -> [выход, закрытие done]
//	     |
//	     v да
//	[nats.Connect] -> [ошибка?]
//	     +-- да -> [проверка отмены, лог ошибки] -> [пауза 2 с]
//	     +-- нет -> [проверка отмены, Store(conn)] -> [ожидание]
//	                                                  +-- closed -> Store(nil) -> [пауза 2 с]
//	                                                  +-- отмена -> Close, Store(nil) -> выход
//	[пауза 2 с] -> [следующая попытка либо выход при отмене]
func (NatsClient *NatsClient) Run(Context context.Context) {
	defer close(NatsClient.Done)
	const RetryDelay = 2 * time.Second
	Logger := NatsClient.Logger.With("operation", "connect_nats")

	for Context.Err() == nil {
		// Этап 1: отдельный сигнал closed отличает окончательное закрытие
		// клиента от временного разрыва, который восстанавливает сама библиотека.
		Closed := make(chan struct{})
		ConnectContext, Span := otel.Tracer("trading_bot/internal/di/transport/nats").Start(Context, "NatsClient.Connect")
		Logger.DebugContext(ConnectContext, "Создание клиента NATS")
		Connection, ConnectError := nats.Connect(
			NatsClient.BootstrapSettings.NatsUrl,
			nats.UserInfo(NatsClient.BootstrapSettings.NatsUser, NatsClient.BootstrapSettings.NatsPassword),
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(-1),
			nats.ReconnectWait(RetryDelay),
			nats.ReconnectJitter(500*time.Millisecond, time.Second),
			nats.PingInterval(10*time.Second),
			nats.MaxPingsOutstanding(2),
			nats.DisconnectErrHandler(func(_ *nats.Conn, DisconnectError error) {
				if Context.Err() == nil && DisconnectError != nil {
					Logger.WarnContext(Context, "Соединение с NATS потеряно", "error", safeConnectionError(DisconnectError))
				}
			}),
			nats.ReconnectHandler(func(_ *nats.Conn) {
				if Context.Err() == nil {
					Logger.InfoContext(Context, "Соединение с NATS установлено")
				}
			}),
			nats.ClosedHandler(func(ClosedConnection *nats.Conn) {
				if Context.Err() == nil {
					Logger.WarnContext(Context, "Соединение с NATS закрыто, клиент будет создан заново", "error", safeConnectionError(ClosedConnection.LastError()))
				}
				close(Closed)
			}),
		)
		// Этап 2: ошибка создания не останавливает приложение. После успеха
		// проверяем отмену до сохранения клиента, чтобы закрыть ненужный ресурс.
		if ConnectError != nil {
			SafeError := safeConnectionError(ConnectError)
			Span.RecordError(SafeError)
			Span.SetStatus(codes.Error, SafeError.Error())
			if Context.Err() == nil {
				Logger.ErrorContext(ConnectContext, "Не удалось создать клиент NATS", "error", SafeError, "retry_after", RetryDelay)
			}
			Span.End()
			if Context.Err() != nil {
				return
			}
		} else {
			if Context.Err() != nil {
				Span.End()
				Connection.Close()
				return
			}
			NatsClient.Connection.Store(Connection)
			Logger.InfoContext(ConnectContext, "Клиент NATS запущен", "status", Connection.Status().String())
			Span.End()
			select {
			case <-Context.Done():
				Connection.Close()
				NatsClient.Connection.Store(nil)
				return
			case <-Closed:
				NatsClient.Connection.Store(nil)
			}
		}

		// Этап 3: ограничиваем частоту пересоздания клиента; отмена не ждёт таймера.
		Timer := time.NewTimer(RetryDelay)
		select {
		case <-Context.Done():
			Timer.Stop()
			return
		case <-Timer.C:
		}
	}
}

// safeConnectionError скрывает адрес с учётными данными в ошибках разбора URL.
func safeConnectionError(ConnectionError error) error {
	var URLError *url.Error
	if errors.As(ConnectionError, &URLError) {
		return errors.New("некорректный URL подключения NATS")
	}
	return ConnectionError
}

// WaitConnection ждёт подключённый клиент, не создавая нового соединения.
// Context отменяет ожидание; результат — текущий *nats.Conn либо Context.Err().
// Соединение может разорваться сразу после возврата: вызывающий код должен
// обрабатывать ошибки последующих операций.
//
// Схема:
//
//	[ctx активен?] -> [Клиент подключён?] -- да -> Conn
//	       |                 |
//	      нет               нет -> [Пауза 100 мс] -> повтор
//	       v
//	    Context.Err()
func (NatsClient *NatsClient) WaitConnection(Context context.Context) (Connection *nats.Conn, WaitError error) {
	Context, Span := otel.Tracer("trading_bot/internal/di/transport/nats").Start(Context, "NatsClient.WaitConnection")
	defer internal.FinishSpan(Span, &WaitError)
	NatsClient.Logger.DebugContext(Context, "Ожидание подключённого клиента NATS", "operation", "wait_connection")
	Ticker := time.NewTicker(100 * time.Millisecond)
	defer Ticker.Stop()

	for {
		if ContextError := Context.Err(); ContextError != nil {
			return nil, ContextError
		}
		if Connection := NatsClient.Connection.Load(); Connection != nil && Connection.IsConnected() {
			return Connection, nil
		}
		// Ожидание очередной проверки прерывается отменой контекста.
		select {
		case <-Context.Done():
			return nil, Context.Err()
		case <-Ticker.C:
		}
	}
}
