BINARY_DIR=bin
BINARY_NAME=server
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)
DOCKER_IMAGE=fts-engine
DOCKER_IMAGE_DEV=fts-engine-dev
DOCKER_PORT=3000
WATCH_PORT=3001
COVERAGE_PROFILE=cover.out

.PHONY: vendor

all: lint test build

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
		--build.cmd "go build -race -o $(BINARY_PATH) ./src" \
		--build.bin "./$(BINARY_PATH) -p $(WATCH_PORT)"

# Test:
test:
	@go test -v -race ./...

coverage:
	@go test -cover -covermode=count -coverprofile=$(COVERAGE_PROFILE) ./...
	@go tool cover -func $(COVERAGE_PROFILE)

# Lint
lint:
	@golangci-lint run --timeout=3m

# Docker:
docker-dev-build:
	@docker build -t $(DOCKER_IMAGE_DEV) --target dev .

docker-check:
	@docker run --rm -v $(shell pwd):/app $(DOCKER_IMAGE_DEV)

docker-test:
	@docker run --rm -v $(shell pwd):/app $(DOCKER_IMAGE_DEV) test

docker-build:
	@docker build -t $(DOCKER_IMAGE) --target prod .

docker-run:
	@docker run --rm --name $(DOCKER_IMAGE) -d -p $(DOCKER_PORT):$(DOCKER_PORT) $(DOCKER_IMAGE) -p $(DOCKER_PORT)

docker-stop:
	@docker stop $(DOCKER_IMAGE)
