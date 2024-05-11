FROM golang:alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd rpc/service/publish && sh build.sh

FROM ubuntu:jammy

RUN sed -i 's/ports.ubuntu.com/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends --fix-missing \
    ffmpeg 

COPY --from=builder /douyin/rpc/service/publish/output/bin/publish /

ENTRYPOINT ["/publish"]
