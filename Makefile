SHELL := bash

run:
	go run -race ./main.go --conf=./test_config.json

build:
	go build
