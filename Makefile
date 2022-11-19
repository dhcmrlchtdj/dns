SHELL := bash

GOFLAGS := \
	-trimpath \
	-buildmode=pie \
	-buildvcs=false \
	-ldflags='-s -w -linkmode=external'

build:
	go build $(GOFLAGS) -o ./_build/app

clean:
	# go clean -testcache ./...
	-rm -rf ./_build

fmt:
	gofumpt -w .
	goimports -w .

lint:
	golangci-lint run

dev:
	go run -race ./main.go --conf=./aur/config.json --log-level=trace --port=1053

upgrade:
	go get -v -t -u ./...
	go mod tidy -v
