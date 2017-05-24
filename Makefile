BINARY=dkconf

EXAMPLES=$(shell find examples/* -type d -exec sh -c '(ls -p "{}"|grep />/dev/null)||echo "{}"' \;)

.PHONY: all examples

all:
	go build -o ${BINARY}-osx main.go
	env GOOS=linux GOARCH=amd64 go build -o ${BINARY}-linux main.go

test:
	go test  -v ./...

get:
	go get

examples:
	-@$(foreach test,$(EXAMPLES),(echo "\033[0;33mdkconf < ${test}\033[0m" && (source $(test)/.env && echo "\033[0;31m\c" && go run main.go -p TEST -s $(test)/template.tmpl | diff $(test)/expected.txt -) ; echo "\033[0m\c"); )

examples-linux:
	-@$(foreach test,$(EXAMPLES),(echo "\033[0;33mdkconf < ${test}\033[0m" && (source $(test)/.env && echo "\033[0;31m\c" && ./dkconf-linux -p TEST -s $(test)/template.tmpl | diff $(test)/expected.txt -) ; echo "\033[0m\c"); )

examples-osx:
	-@$(foreach test,$(EXAMPLES),(echo "\033[0;33mdkconf < ${test}\033[0m" && (source $(test)/.env && echo "\033[0;31m\c" && ./dkconf-osx -p TEST -s $(test)/template.tmpl | diff $(test)/expected.txt -) ; echo "\033[0m\c"); )
