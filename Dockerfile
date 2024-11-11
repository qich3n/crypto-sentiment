FROM golang:1.21-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN go build -o main ./cmd

# Expose port
EXPOSE 10000

# Run
CMD ["./main"]