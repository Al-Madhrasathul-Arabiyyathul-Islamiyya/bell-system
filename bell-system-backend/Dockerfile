# Build stage
FROM golang:1.26-alpine AS builder

ARG VERSION=dev

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=${VERSION}" -o /bin/bell-system-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/bell-system-migrate ./cmd/migrate

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates curl

# Trust the organisation's internal ADCS CA
COPY adcs-ca.pem /usr/local/share/ca-certificates/adcs-ca.crt
RUN update-ca-certificates

WORKDIR /app

COPY --from=builder /bin/bell-system-server /app/bell-system-server
COPY --from=builder /bin/bell-system-migrate /app/bell-system-migrate
COPY --from=builder /src/migrations /app/migrations

RUN mkdir -p /app/data/audio

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/status || exit 1

ENTRYPOINT ["/app/bell-system-server"]
