FROM golang:alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd src/service/message && sh build.sh

FROM scratch

ENV TZ=Asia/Shanghai

COPY --from=builder /douyin/src/service/message/output/bin/message /

ENTRYPOINT ["/message"]
