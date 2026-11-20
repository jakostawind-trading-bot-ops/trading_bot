# syntax=docker/dockerfile:1
FROM golang:1.27.3-alpine AS base
WORKDIR /app
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY go.mod go.sum ./
RUN go mod download

FROM base AS dev
RUN go install github.com/air-verse/air@v1.67.4
COPY cmd/ ./cmd/
COPY internal/ ./internal/
CMD ["air", "--build.cmd", "go build -o ./tmp/trading_bot ./cmd/trading_bot", "--build.entrypoint", "./tmp/trading_bot", "--build.send_interrupt=true"]

FROM base AS build
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/trading_bot ./cmd/trading_bot

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=build /out/trading_bot /trading_bot
ENTRYPOINT ["/trading_bot"]
