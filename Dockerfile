# Multi-stage build for dashai-go
#
# go-factory-io module path is github.com/dashfactory/go-factory-io
# but the repo lives at github.com/seikaikyo/go-factory-io.
# We use git insteadOf to redirect the module fetch.

FROM golang:1.26-bookworm AS builder

RUN git config --global url."https://github.com/seikaikyo/go-factory-io".insteadOf "https://github.com/dashfactory/go-factory-io"

ENV GONOSUMCHECK=github.com/dashfactory/go-factory-io
ENV GOFLAGS=-mod=mod

WORKDIR /app
COPY go.mod go.sum ./

# Remove local replace directive for Docker build
RUN sed -i '/^replace.*go-factory-io/d' go.mod
RUN go mod tidy && go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server/

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8101
CMD ["./server"]
