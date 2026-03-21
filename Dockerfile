# Stage 1: Build Go binary
FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /squid-helper ./cmd/squid-helper

# Stage 2: Runtime with Squid
FROM alpine:3.21

RUN apk add --no-cache squid

COPY --from=builder /squid-helper /usr/local/bin/squid-helper
COPY squid/squid.conf /etc/squid/squid.conf
COPY squid/ERR_TIME_LIMIT /usr/share/squid/errors/ja/ERR_TIME_LIMIT
COPY squid/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh && \
    mkdir -p /var/lib/squid-brocker /etc/squid-brocker

VOLUME ["/var/lib/squid-brocker"]
VOLUME ["/etc/squid-brocker"]

EXPOSE 3128

# Squid spawns the Go helper as a child process via external_acl_type
ENTRYPOINT ["/entrypoint.sh"]
