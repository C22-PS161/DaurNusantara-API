# syntax=docker/dockerfile:1

FROM golang:1.18-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o /server

FROM alpine:3

WORKDIR /
COPY --from=builder /server /server
COPY sample_data.json /data.json

RUN apk add --no-cache ca-certificates

CMD ["/server"]