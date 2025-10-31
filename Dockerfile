# syntax=docker/dockerfile:1.7

FROM golang:1.25 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o /workspace/bin/pinguin ./cmd/server

FROM gcr.io/distroless/base-debian12

COPY --from=builder /workspace/bin/pinguin /usr/local/bin/pinguin

EXPOSE 50051

ENTRYPOINT ["/usr/local/bin/pinguin"]
