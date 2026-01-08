# Build stage: Go
FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

# Run stage
FROM alpine:3.18

WORKDIR /app

# Install system dependencies
RUN apk add --no-cache \
    curl \
    ca-certificates

# Copy binary from builder
COPY --from=builder /app/main .

# Copy classification rules (just in case we use them in Go later, 
# although logic is currently hardcoded in internal/processor/common/common.go)
COPY ./classification_rules.json .

# Create directory for classified data
RUN mkdir -p /app/DatosClasificados

EXPOSE 8080

CMD ["./main"]
