FROM golang:alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd rpc/service/message && sh build.sh

FROM scratch

COPY --from=builder /douyin/rpc/service/message/output/bin/message /

ENTRYPOINT ["/message"]
