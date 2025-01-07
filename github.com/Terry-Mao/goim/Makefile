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

logic:
	nohup target/logic${prifix} -conf=target/logic.toml 2>&1 > target/logic.log &
comet:
	nohup target/comet${prifix} -conf=target/comet.toml 2>&1 > target/comet.log &
job:
	nohup target/job${prifix} -conf=target/job.toml  2>&1 > target/job.log &
stop:
	pkill -f target/logic
	pkill -f target/job
	pkill -f target/comet