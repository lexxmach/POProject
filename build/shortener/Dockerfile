# syntax=docker/dockerfile:1

FROM golang:1.22-alpine as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o shortener cmd/shortener/main.go

FROM alpine:3.16

WORKDIR /app

COPY --from=builder /build/shortener ./
COPY ./configs/shortener.json ./config.json

CMD ./shortener -c config.json

EXPOSE 8000