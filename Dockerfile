ARG GO_VERSION=1.25.3

FROM golang:1.22-bookworm AS builder

ARG GO_VERSION

RUN apt-get update -y \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*

# Install the required Go toolchain version manually to avoid relying on
# automatic downloads which are blocked in some environments.
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf /tmp/go.tgz \
    && rm -f /tmp/go.tgz

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH="$GOROOT/bin:$GOPATH/bin:$PATH"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o agent ./cmd/agent

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/agent /usr/local/bin/agent
COPY config ./config

ENV GIN_MODE=release

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/agent"]