BINARY_DIR=bin
BINARY_NAME=server
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)
DOCKER_IMAGE=fts-engine-rest
COVERAGE_PROFILE=cover.out

.PHONY: vendor

all: lint test

# Build:
get:
	@go get ./src

build:
	@go build -o $(BINARY_PATH) ./src

clean:
	@go clean --cache
	@rm -rf $(BINARY_DIR)
	@rm $(COVERAGE_PROFILE)

vendor:
	@GO111MODULE=on go mod vendor

watch:
	$(eval PACKAGE_NAME=$(shell head -n 1 go.mod | cut -d ' ' -f2))
	@docker run -it --rm -w /go/src/$(PACKAGE_NAME) -v $(shell pwd):/go/src/$(PACKAGE_NAME) -p 8080:8080 cosmtrek/air --build.cmd "make build" --build.bin "./$(BINARY_PATH)"

# Test:
test:
	@go test -v -race ./...

coverage:
	@go test -cover -covermode=count -coverprofile=$(COVERAGE_PROFILE) ./...
	@go tool cover -func $(COVERAGE_PROFILE)

# Lint
lint:
	@golangci-lint run

# Docker:
docker-build-run:
	@docker build -t $(DOCKER_IMAGE) .
	@docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

docker-run:
	@docker run --rm -p 8080:8080 $(DOCKER_IMAGE)
