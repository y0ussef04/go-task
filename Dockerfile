# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Initialize module and download dependencies
COPY main.go ./
COPY internal/ ./internal/
COPY cmd/ ./cmd/

RUN go mod init dbmanager && \
    go get github.com/go-sql-driver/mysql && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o dbmanager . && \
    CGO_ENABLED=0 GOOS=linux go build -o dbcli ./cmd/cli

# Stage 2: Run
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/dbmanager .
COPY --from=builder /app/dbcli .
COPY static/ ./static/

EXPOSE 8081

CMD ["./dbmanager"]