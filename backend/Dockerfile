FROM golang:1.20

WORKDIR /app

# Install dependencies
RUN apt-get update && apt-get install -y gcc libc6-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o vpn-service .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./vpn-service"]
