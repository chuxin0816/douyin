FROM golang:alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd src/rpc/service/publish && sh build.sh

FROM alpine:latest

ENV TZ=Asia/Shanghai

RUN set -eux; \
    apk update; \
    apk add --no-cache ffmpeg

COPY --from=builder /douyin/rpc/service/publish/output/bin/publish /

ENTRYPOINT ["/publish"]