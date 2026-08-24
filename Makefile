APP_NAME := strata-log
CMD := ./cmd/strata-log
BINARY := bin/$(APP_NAME)

.PHONY: all build run test race vet fmt check benchmark clean \
	docker-build docker-up docker-down

all: check

build:
	@mkdir -p bin
	go build -o $(BINARY) $(CMD)

run:
	go run $(CMD)

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

check: fmt test race vet

benchmark:
	go test -bench=. -benchmem ./benchmarks/...

docker-build:
	docker build -f deployments/Dockerfile -t $(APP_NAME):latest .

docker-up:
	docker compose -f deployments/docker-compose.yml up --build

docker-down:
	docker compose -f deployments/docker-compose.yml down

clean:
	rm -rf bin
