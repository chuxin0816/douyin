FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd rpc/service/user && sh build.sh

FROM scratch

COPY --from=builder /douyin/rpc/service/user/output/bin/user /

ENTRYPOINT ["/user"]