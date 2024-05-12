FROM golang:alpine AS builder

WORKDIR /douyin

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o app api/main.go

FROM scratch

ENV TZ=Asia/Shanghai

COPY --from=builder /douyin/app /

ENTRYPOINT ["/app"]