.PHONY: gotool build clean

BINARY="douyin"

all: gotool	build

gotool:
	go mod tidy
	go fmt ./
	go vet ./

build:
	go build -o $(BINARY)

run:
	go run main.go

clean: 
	if [ -f $(BINARY) ] ; then rm $(BINARY) ; fi
