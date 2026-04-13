
FROM golang:1.22-alpine

WORKDIR /app

COPY main.go .
COPY static/ ./static/

RUN go mod init realestate-mgmt && \
    go get github.com/go-sql-driver/mysql && \
    go mod tidy && \
    go build -o realestate-mgmt .

CMD ["./realestate-mgmt"]