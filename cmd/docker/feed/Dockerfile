FROM golang:alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd rpc/service/feed && sh build.sh

FROM scratch

COPY --from=builder /douyin/rpc/service/feed/output/bin/feed /

ENTRYPOINT ["/feed"]
