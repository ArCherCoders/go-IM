# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
basepath=github.com\Terry-Mao\goim\api

api:
	protoc --proto_path=./ \
	--go-grpc_out=require_unimplemented_servers=false:. \
	--go_out=paths=source_relative:. \
	${basepath}\comet\comet.proto \
	${basepath}\logic\logic.proto \
    ${basepath}\protocol\protocol.proto \
    ${basepath}\server\Hello.proto

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

