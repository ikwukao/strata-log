.PHONY: help fmt test test-race vet build run benchmark check \
	docker-build up down restart logs health metrics \
	clean

# -----------------------------------------------------------------------------
# Configuration
# -----------------------------------------------------------------------------

APP_NAME := strata-log
BINARY := bin/$(APP_NAME)
CMD := ./cmd/strata-log
COMPOSE := docker compose -f deployments/docker-compose.yml
PORT := 9090

# -----------------------------------------------------------------------------
# Help
# -----------------------------------------------------------------------------

help:
	@echo "Strata-Log development commands:"
	@echo ""
	@echo "  make fmt           Format Go source files"
	@echo "  make test          Run all tests"
	@echo "  make test-race     Run tests with race detection"
	@echo "  make vet           Run go vet"
	@echo "  make build         Build the Strata-Log binary"
	@echo "  make run           Run Strata-Log locally"
	@echo "  make benchmark     Run project benchmarks"
	@echo "  make check         Run formatting, tests, race tests, vet, and build"
	@echo ""
	@echo "  make docker-build  Build the Docker image"
	@echo "  make up            Start the Docker Compose stack"
	@echo "  make down          Stop the Docker Compose stack"
	@echo "  make restart       Restart the Docker Compose stack"
	@echo "  make logs          Follow Strata-Log container logs"
	@echo ""
	@echo "  make health        Check the service health endpoint"
	@echo "  make metrics       Display Prometheus metrics"
	@echo ""
	@echo "  make clean         Remove local build artifacts"

# -----------------------------------------------------------------------------
# Development
# -----------------------------------------------------------------------------

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags="-s -w" \
		-o $(BINARY) \
		$(CMD)

run:
	go run $(CMD)

benchmark:
	go test -bench=. -benchmem ./benchmarks/...

check: fmt test test-race vet build

# -----------------------------------------------------------------------------
# Docker
# -----------------------------------------------------------------------------

docker-build:
	$(COMPOSE) build

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) down
	$(COMPOSE) up -d

logs:
	$(COMPOSE) logs -f strata-log

# -----------------------------------------------------------------------------
# Operations
# -----------------------------------------------------------------------------

health:
	curl -fsS http://localhost:$(PORT)/healthz
	@echo

metrics:
	curl -fsS http://localhost:$(PORT)/metrics

# -----------------------------------------------------------------------------
# Cleanup
# -----------------------------------------------------------------------------

clean:
	rm -rf bin
	rm -f coverage.out
