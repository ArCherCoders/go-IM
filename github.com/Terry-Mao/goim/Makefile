# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOHOSTOS:=$(shell go env GOHOSTOS)
prifix=
ifeq ($(GOHOSTOS), windows)
	prifix=.exe
endif

build:
	$(GOBUILD) -o target/comet${prifix} cmd/comet/main.go
	$(GOBUILD) -o target/logic${prifix} cmd/logic/main.go
	$(GOBUILD) -o target/job${prifix} cmd/job/main.go

run:
	target/logic${prifix} -conf=target/logic.toml
	target/comet${prifix} -conf=target/comet.toml
	target/job${prifix} -conf=target/job.toml