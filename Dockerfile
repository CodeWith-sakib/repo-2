# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Copy module manifests first for dependency caching
COPY go.mod ./
RUN go mod download || true

# Copy source files
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/kestrel ./cmd/kestrel

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/kestrel /usr/local/bin/kestrel

EXPOSE 8080

ENTRYPOINT ["kestrel"]
CMD ["server", "-host=0.0.0.0", "-port=8080"]
