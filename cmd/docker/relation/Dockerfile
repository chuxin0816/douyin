FROM golang:1.23.0-alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN cd src/service/relation && sh build.sh

FROM scratch

ENV TZ=Asia/Shanghai

COPY --from=builder /douyin/src/service/relation/output/bin/relation /

ENTRYPOINT ["/relation"]
