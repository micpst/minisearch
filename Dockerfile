# syntax = docker/dockerfile:1
FROM golang:1.19-buster AS base

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM golangci/golangci-lint:v1.50.1 AS dev

WORKDIR /app

COPY --from=base /go/pkg/mod /go/pkg/mod

ENTRYPOINT ["make"]

FROM base AS build

COPY src ./src
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o ./bin/server ./src

FROM gcr.io/distroless/base-debian11 AS prod

COPY --from=build /app/bin/server /

USER nonroot:nonroot

ENTRYPOINT ["/server"]
