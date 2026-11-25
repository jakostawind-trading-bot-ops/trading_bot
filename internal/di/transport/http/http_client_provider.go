package di_transport_http

import (
	"net/http"
	"time"
)

// HTTPClientProvider возвращает HTTP-клиент с таймаутом 10 секунд на запрос.
func HTTPClientProvider() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}
