.PHONY: dev build test migrate

dev:
	air

build:
	go build -o bin/server.exe ./cmd/server

test:
	go test ./...

migrate:
	go run ./cmd/migrate

clean:
	rm -rf ./tmp ./bin
