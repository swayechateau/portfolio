# Use the official Golang image as a base image
FROM golang:latest AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
ENV CGO_ENABLED=0
RUN go build -o portfolio main.go

# Start a new stage from scratch
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/portfolio .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/storage ./storage

# Expose port 5050 to the outside world
EXPOSE 5050

# Command to run the executable
CMD ["./portfolio"]
