BINARY_DIR=bin
BINARY_NAME=server
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)
RELEASE_NAME=fts-engine
COVERAGE_PROFILE=cover.out
DOCKER_IMAGE=fts-engine
DOCKER_IMAGE_DEV=$(DOCKER_IMAGE)-dev
DOCKER_IMAGE_WATCH=$(DOCKER_IMAGE)-watch
DOCKER_PORT=3000
WATCH_PORT=3001

.PHONY: vendor

all: lint coverage build

# Build:
setup:
	@go mod download

build:
	@go build -o $(BINARY_PATH) ./cmd/server

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
		--name $(DOCKER_IMAGE_WATCH) \
		cosmtrek/air \
		--build.cmd "go build -race -o $(BINARY_PATH) ./cmd/server" \
		--build.bin "./$(BINARY_PATH) -p $(WATCH_PORT)"

# Test:
test:
	@go test -v -race ./...

coverage:
	@go test -v -race -cover -covermode=atomic -coverprofile=$(COVERAGE_PROFILE) ./...
	@go tool cover -func $(COVERAGE_PROFILE)

# Benchmark:
bench:
	@go test -bench=. -run=^a -benchtime=5x ./...

# Lint
lint:
	@golangci-lint run --timeout=3m

# Release:
release:
	@go build -o $(BINARY_NAME) ./cmd/server
	@tar -czvf $(RELEASE_NAME).tar.gz $(BINARY_NAME) README.md LICENSE.md
	@rm $(BINARY_NAME)

# Docker dev:
docker-dev-setup:
	@docker build -t $(DOCKER_IMAGE_DEV) --target dev .
	@docker run --rm --name $(DOCKER_IMAGE_DEV) -d -v $(shell pwd):/app $(DOCKER_IMAGE_DEV)

docker-dev-stop:
	@docker stop $(DOCKER_IMAGE_DEV)

docker-check:
	@docker exec $(DOCKER_IMAGE_DEV) make

docker-test:
	@docker exec $(DOCKER_IMAGE_DEV) make test

docker-bench:
	@docker exec $(DOCKER_IMAGE_DEV) make bench

# Docker prod:
docker-build:
	@docker build -t $(DOCKER_IMAGE) --target release .

docker-run:
	@docker run --rm --name $(DOCKER_IMAGE) -d -p $(DOCKER_PORT):$(DOCKER_PORT) $(DOCKER_IMAGE) -p $(DOCKER_PORT)

docker-stop:
	@docker stop $(DOCKER_IMAGE)
