.PHONY: build run test clean lint

APP_NAME := agenthub
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v -cover

clean:
	rm -rf $(BUILD_DIR)

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
