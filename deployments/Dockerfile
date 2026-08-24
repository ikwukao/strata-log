FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/strata-log ./cmd/strata-log

FROM alpine:3.22

RUN addgroup -S strata && \
    adduser -S -G strata strata && \
    mkdir -p /data && \
    chown -R strata:strata /data

WORKDIR /app

COPY --from=builder /out/strata-log /app/strata-log

USER strata

EXPOSE 9090

ENV STRATA_LOG_HOST=0.0.0.0 \
    STRATA_LOG_PORT=9090 \
    STRATA_LOG_STORAGE_PATH=/data/strata-log.db

ENTRYPOINT ["/app/strata-log"]
