FROM golang:1.23.0-alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o app src/service/api/main.go

FROM scratch

ENV TZ=Asia/Shanghai

COPY --from=builder /douyin/app /

ENTRYPOINT ["/app"]