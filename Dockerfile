# syntax = docker/dockerfile:1
FROM golang:1.19-buster AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY src ./src
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o ./bin/server ./src

FROM gcr.io/distroless/base-debian11

COPY --from=build /app/bin/server /

USER nonroot:nonroot

ENTRYPOINT ["/server"]
