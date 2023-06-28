# Use the official Golang image as the base image
FROM golang:1.17-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod ./
RUN go mod download

# Copy the server source code into the container
COPY . .

# Build the server binary
RUN go build -o server

# Expose the server port
EXPOSE 9090

# Start the server
CMD ["./server"]