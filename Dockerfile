FROM golang:1.21 AS builder

WORKDIR /app

COPY . ./

ENV CGO_ENABLED=0
RUN go build -o fileb

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:buster-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/fileb /fileb

ENTRYPOINT [ "/fileb" ]