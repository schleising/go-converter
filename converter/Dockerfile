# Start with the golang image
FROM golang:latest as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY src/* /app/

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o go-converter .

# New stage using the ffmpeg image
FROM linuxserver/ffmpeg:latest

# Update and upgrade the image
RUN apt update && apt upgrade -y

# Create a directory to store the converted files
RUN mkdir /Conversions

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/go-converter /app/go-converter

# Command to run the executable
ENTRYPOINT [ "/app/go-converter" ]
