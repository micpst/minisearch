BINARY_DIR=bin
BINARY_NAME=server
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)
COVERAGE_PROFILE=cover.out
DOCKER_IMAGE=fts-engine
DOCKER_IMAGE_DEV=fts-engine-dev
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
		cosmtrek/air \
		--build.cmd "go build -race -o $(BINARY_PATH) ./cmd/server" \
		--build.bin "./$(BINARY_PATH) -p $(WATCH_PORT)"

# Test:
test:
	@go test -v -race ./...

coverage:
	@go test -v -race -cover -covermode=atomic -coverprofile=$(COVERAGE_PROFILE) ./...
	@go tool cover -func $(COVERAGE_PROFILE)

# Lint
lint:
	@golangci-lint run --timeout=3m

# Docker dev:
docker-dev-setup:
	@docker build -t $(DOCKER_IMAGE_DEV) --target dev .
	@docker run --rm --name $(DOCKER_IMAGE_DEV) -v $(shell pwd):/app $(DOCKER_IMAGE_DEV)

docker-dev-stop:
	@docker stop $(DOCKER_IMAGE_DEV)

docker-check:
	@docker exec $(DOCKER_IMAGE_DEV) make

docker-test:
	@docker exec $(DOCKER_IMAGE_DEV) make test

# Docker prod:
docker-build:
	@docker build -t $(DOCKER_IMAGE) --target prod .

docker-run:
	@docker run --rm --name $(DOCKER_IMAGE) -d -p $(DOCKER_PORT):$(DOCKER_PORT) $(DOCKER_IMAGE) -p $(DOCKER_PORT)

docker-stop:
	@docker stop $(DOCKER_IMAGE)
