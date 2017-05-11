BINARY=dkconf

PHONY: all

test:
	go test  -v ./...

get:
	go get

all:
	go build -o ${BINARY}-osx main.go
	env GOOS=linux GOARCH=amd64 go build -o ${BINARY}-linux main.go
