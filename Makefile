SHELL := bash

run:
	go run -race ./main.go --log-level=trace --conf=./aur/config.json

.PHONY: build
build:
	@mkdir -p build
	go build -o build .
