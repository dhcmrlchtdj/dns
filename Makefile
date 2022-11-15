SHELL := bash

GOFLAGS := \
	-trimpath \
	-buildmode=pie \
	-buildvcs=false \
	-ldflags='-s -w -linkmode=external'

run:
	go run -race ./main.go --conf=./aur/config.json --log-level=trace --port=1053

.PHONY: build
build:
	@mkdir -p build
	go build $(GOFLAGS) -o build .

fmt:
	gofumpt -w .
	goimports -w .
