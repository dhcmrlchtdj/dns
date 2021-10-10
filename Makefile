SHELL := bash

run:
	go run -race ./main.go --conf=./test_config.json

build:
	mkdir build
	go build -o build .
