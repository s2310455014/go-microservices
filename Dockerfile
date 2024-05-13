# Pull in the layer of the base image: alpine:3.10.3
FROM golang:1.22-alpine

LABEL maintainer="s2310455014@fhooe.at"

WORKDIR /src

COPY go.mod go.sum ./
COPY *.go ./

ENV CGO_ENABLED=0

# List files in /src for debugging purposes
RUN ls /src

# Build the application
RUN go build -o myapp .

# Expose port 8888 for the application
EXPOSE 8010

# Command to run the executable
CMD ["./myapp"]