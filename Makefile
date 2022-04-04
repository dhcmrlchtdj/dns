SHELL := bash

run:
	go run -race ./main.go --conf=./aur/config.json --log-level=trace --port=1053

.PHONY: build
build:
	@mkdir -p build
	go build -o build .
