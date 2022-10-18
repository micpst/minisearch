BINARY_DIR=bin
BINARY_NAME=server
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)

all: build test

install:
	go get ./src

lint:
	golangci-lint run

build:
	go build -o $(BINARY_PATH) ./src

test:
	go test -v -race ./...

run:
	go build -o $(BINARY_PATH) ./src
	./$(BINARY_PATH)

clean:
	go clean
	rm -rf $(BINARY_DIR)
