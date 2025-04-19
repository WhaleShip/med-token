FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

EXPOSE 8080

RUN go mod download
RUN go build -o app ./cmd/app/main.go