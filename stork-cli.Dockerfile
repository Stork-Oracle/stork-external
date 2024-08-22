FROM golang:1.22-alpine AS build

WORKDIR /app/cli

COPY cli/go.mod cli/go.sum ./
RUN go mod download

# Copy the source code from the cli directory into the container
COPY cli/ .

RUN GOOS=linux GOARCH=arm64 go build -o /stork .

FROM alpine:3.18

COPY --from=build /stork /usr/local/bin/stork

ENTRYPOINT ["/usr/local/bin/stork"]