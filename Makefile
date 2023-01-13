SHELL := bash
.SHELLFLAGS = -O globstar -c

GOFLAGS := \
	-trimpath \
	-buildvcs=false \
	-buildmode=pie \
	-ldflags='-s -w -linkmode=external'

###

.PHONY: dev build fmt lint test clean outdated upgrade

dev:
	go run -race ./main.go --conf=./aur/config.json --log-level=trace --port=1053

build:
	go build $(GOFLAGS) -o ./_build/app

fmt:
	gofumpt -w .
	goimports -w .

lint:
	golangci-lint run

test:
	ENV=test TZ=UTC go test -race ./...

clean:
	go clean -testcache ./...
	-rm -rf ./_build

# outdated:
#     go list -u -m -f '{{if not .Indirect}}{{.}}{{end}}' all

upgrade:
	go get -v -t -u ./...
	go mod tidy -v
