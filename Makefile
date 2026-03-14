.PHONY: run build test test-cover vet docker clean

run:
	go run ./cmd

build:
	CGO_ENABLED=1 go build -o asistente ./cmd

test:
	go test -race ./...

test-cover:
	go test -race -cover -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

vet:
	go vet ./...

docker:
	docker compose up -d --build asistente

docker-all:
	docker compose up -d --build

docker-down:
	docker compose down

clean:
	rm -f asistente cmd.exe coverage.out
