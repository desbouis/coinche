# https://hub.docker.com/_/golang
FROM golang:1.19 as builder

WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

# Copy local code to the builder image.
COPY . ./

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o bin/coinche

# https://hub.docker.com/_/alpine
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3.17.0
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the binary and static files from the builder stage.
COPY --from=builder /app/bin/coinche ./
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/assets ./assets

EXPOSE 8080

# Run server on container startup.
CMD ["/app/coinche"]
