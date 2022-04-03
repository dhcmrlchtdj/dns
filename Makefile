SHELL := bash

run:
	go run -race ./main.go --conf=./aur/config.json

.PHONY: build
build:
	@mkdir -p build
	go build -o build .
