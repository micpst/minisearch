# syntax = docker/dockerfile:1
FROM golang:1.20-buster AS base
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM golangci/golangci-lint:v1.51 AS dev
WORKDIR /app
COPY --from=base /go/pkg/mod /go/pkg/mod
ENTRYPOINT ["tail", "-f", "/dev/null"]

FROM base AS build
COPY api ./api
COPY cmd ./cmd
COPY pkg ./pkg
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o ./bin/server ./cmd/server

FROM gcr.io/distroless/base-debian11 AS release
COPY --from=build /app/bin/server /
ENV GIN_MODE=release
USER nonroot:nonroot
ENTRYPOINT ["/server"]
