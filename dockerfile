FROM golang:latest AS builder

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

FROM ubuntu:lunar

COPY ./wait-for.sh /
COPY ./config/config.yaml /config/config.yaml

COPY --from=builder /build/douyin /

# x86架构
# RUN echo "deb http://mirrors.tuna.tsinghua.edu.cn/ubuntu/ lunar main restricted universe multiverse" > /etc/apt/sources.list && \
#     echo "deb http://mirrors.tuna.tsinghua.edu.cn/ubuntu/ lunar-updates main restricted universe multiverse" >> /etc/apt/sources.list && \
#     echo "deb http://mirrors.tuna.tsinghua.edu.cn/ubuntu/ lunar-backports main restricted universe multiverse" >> /etc/apt/sources.list && \
#     echo "http://security.ubuntu.com/ubuntu/ lunar-security main restricted universe multiverse" >> /etc/apt/sources.list 

# arm架构
RUN echo "deb http://mirrors.tuna.tsinghua.edu.cn/ubuntu-ports/ lunar main restricted universe multiverse" > /etc/apt/sources.list && \
    echo "deb http://mirrors.tuna.tsinghua.edu.cn/ubuntu-ports/ lunar-updates main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.tuna.tsinghua.edu.cn/ubuntu-ports/ lunar-backports main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb http://ports.ubuntu.com/ubuntu-ports/ lunar-security main restricted universe multiverse" >> /etc/apt/sources.list

RUN set -eux; \
    apt-get update; \
    apt-get install -y \
    ffmpeg \
    netcat-traditional; \
    chmod 755 wait-for.sh