FROM golang:alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd rpc/service/favorite && sh build.sh

FROM scratch

COPY --from=builder /douyin/rpc/service/favorite/output/bin/favorite /

ENTRYPOINT ["/favorite"]