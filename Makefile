BINARY_DIR=bin
BINARY_NAME=server
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)

all: build test

install:
	go get ./cmd/server

lint:
	golangci-lint run

build:
	go build -o $(BINARY_PATH) ./cmd/server/

test:
	go test -v -race ./pkg/...

run:
	go build -o $(BINARY_PATH) ./cmd/server/
	./$(BINARY_PATH)

clean:
	go clean
	rm -rf $(BINARY_DIR)
