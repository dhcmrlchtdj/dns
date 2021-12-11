SHELL := bash

run:
	go run -race ./main.go --conf=./test_config.json

.PHONY: build
build:
	@mkdir -p build
	go build -o build .
