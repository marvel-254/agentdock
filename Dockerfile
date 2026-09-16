# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod *.go ./
RUN go build -o orionos .

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/orionos .

RUN mkdir -p /data

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

ENTRYPOINT ["./orionos"]
