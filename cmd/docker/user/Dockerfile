FROM golang:alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd src/rpc/service/user && sh build.sh

FROM scratch

ENV TZ=Asia/Shanghai

COPY --from=builder /douyin/src/rpc/service/user/output/bin/user /

ENTRYPOINT ["/user"]