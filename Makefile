SHELL := bash
.SHELLFLAGS := -O globstar -e -u -o pipefail -c
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables

GOFLAGS := \
	-trimpath \
	-buildvcs=false \
	-buildmode=pie

###

.PHONY: dev build fmt lint test clean outdated upgrade

build:
	CGO_ENABLED=0 go build $(GOFLAGS) -o ./_build/app

dev:
	go run -race ./main.go --conf=./aur/config.json --log-level=trace --port=1053

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
