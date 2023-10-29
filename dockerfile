FROM golang:1.21.2 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o douyin

FROM ubuntu:jammy

COPY ./wait-for.sh /
COPY ./config/config.json /config/config.json

COPY --from=builder /build/douyin /

RUN set -eux; \
    apt-get update; \
    apt-get install -y \
    --no-install-recommends \
    netcat; \
    chmod 755 wait-for.sh
