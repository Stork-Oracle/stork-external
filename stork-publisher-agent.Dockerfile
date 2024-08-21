FROM golang:1.22-alpine AS build

WORKDIR /app/cli

COPY cli/go.mod cli/go.sum ./
RUN go mod download

# Copy the source code from the cli directory into the container
COPY cli/ .

RUN GOOS=linux GOARCH=arm64 go build -o /stork .

# Stage 2: Create a lightweight image with the Cobra CLI
FROM alpine:3.18

# Copy the binary from the build stage
COPY --from=build /stork /usr/local/bin/stork

# Set the entrypoint to your Cobra CLI
ENTRYPOINT ["/usr/local/bin/stork", "publisher-agent"]