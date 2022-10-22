BINARY_DIR=bin
BINARY_NAME=server
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)
DOCKER_IMAGE=fts-engine-rest
DOCKER_PORT=3000
WATCH_PORT=3001
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
	@docker run -it --rm \
		-w /go/src/$(PACKAGE_NAME) \
		-v $(shell pwd):/go/src/$(PACKAGE_NAME) \
		-p $(WATCH_PORT):$(WATCH_PORT) \
		cosmtrek/air \
		--build.cmd "make build" \
		--build.bin "./$(BINARY_PATH) -p $(WATCH_PORT)"

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
	@docker run --rm -p $(DOCKER_PORT):$(DOCKER_PORT) $(DOCKER_IMAGE) -p $(DOCKER_PORT)

docker-run:
	@docker run --rm -p $(DOCKER_PORT):$(DOCKER_PORT) $(DOCKER_IMAGE) -p $(DOCKER_PORT)
