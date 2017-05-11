BINARY=dkconf

PHONY: all

test:
	go test  -v ./...

get:
	go get

all:
	go build -o ${BINARY} main.go
