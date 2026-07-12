# Multi-stage build for Go services
FROM golang:1.23-alpine AS builder

ARG SERVICE_NAME

WORKDIR /src
# Use workspace/monorepo context
COPY . .
WORKDIR /src/services/${SERVICE_NAME}

RUN go mod download
# Build a static binary with no debug info
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/service .

# Use distroless for runtime
FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=builder /app/service /app/service
USER nonroot:nonroot
ENTRYPOINT ["/app/service"]
