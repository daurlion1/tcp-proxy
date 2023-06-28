# Use the official Golang image as the base image
FROM golang:1.17-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the proxy source code into the container
COPY . .

# Build the proxy binary
RUN go build -o proxy

# Start the proxy
CMD ["./proxy"]
