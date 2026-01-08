# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

# Run stage
FROM python:3.9-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Python libraries for normalization
RUN pip install pandas pypdf2 openpyxl

# Copy binary from builder
COPY --from=builder /app/main .

# Copy Python scripts and classification rules
# (Assuming the scripts are in the root directory relative to the backend)
COPY ./normalizer.py .
COPY ./process_transactions.py .
COPY ./classification_rules.json .

EXPOSE 8080

CMD ["./main"]
