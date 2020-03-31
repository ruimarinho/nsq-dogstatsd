# Start from the latest golang base image.
FROM golang AS builder

# Set the current working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum files.
COPY go.mod go.sum ./

# Copy the source from the current directory to the working directory inside the container.
COPY . .

# Build the `nsq_to_dogstatsd` binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/nsq_to_dogstatsd

# Start from scratch image.
FROM scratch

# Copy the static executable.
COPY --from=builder /go/bin/nsq_to_dogstatsd /go/bin/nsq_to_dogstatsd

# Run the `nsq_to_dogstatsd` binary.
ENTRYPOINT ["/go/bin/nsq_to_dogstatsd"]
