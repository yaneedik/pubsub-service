.PHONY: build run proto

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

proto:
	protoc --go_out=. --go-grpc_out=. ./internal/service/service.proto