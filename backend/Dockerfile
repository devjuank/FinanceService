# Build stage: Go
FROM golang:1.20-alpine AS go-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

# Build stage: Python (optional, if we wanted to pre-install dependencies)
# However, we can just install them in the final stage.

# Run stage
FROM python:3.9-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Python libraries for normalization
# pypdf is the modern replacement for pypdf2
RUN pip install --no-cache-dir pandas pypdf openpyxl

# Copy binary from builder
COPY --from=go-builder /app/main .

# Copy Python scripts and classification rules
COPY ./normalizer.py .
COPY ./process_transactions.py .
COPY ./classification_rules.json .

# Create directory for classified data
RUN mkdir -p /app/DatosClasificados

EXPOSE 8080

CMD ["./main"]
