FROM golang:1.21-alpine

WORKDIR /app

# Copy go mod files (check if go.mod and go.sum are changed, if not, the docker will use the cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN go build -o main ./cmd

# Expose the port
EXPOSE 10000

# Run the application
CMD ["./main"]
