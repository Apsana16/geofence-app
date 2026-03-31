# Use official Go image
FROM golang:1.26-alpine

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first (better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all project files
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port
EXPOSE 8081

# Run the application
CMD ["./main"]