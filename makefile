makefile

.PHONY: run test race cover build tidy clean

run:
	go run ./cmd/api

build:
	@mkdir -p bin
	go build -o bin/api ./cmd/api

test:
	go test ./... -count=1

test:
	go test ./... -race -count=1 

cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html


tidy:
	go mod tidy
	go mod verify

clean:
	rm -rf bin coverage.out coverage.html
