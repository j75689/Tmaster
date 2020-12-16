.PHONY: build bench-redis-mq

PROJECT_PATH=$(shell pwd)
GO_TEST_CMD=$(if $(shell which richgo),richgo test,go test)
REPO=github.com/j75689/Tmaster
VERSION=$(shell git symbolic-ref --short HEAD)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_COMMIT_DATE=$(shell git log -n1 --pretty='format:%cd' --date=format:'%Y-%m-%d_%H:%M:%S')

tools:
	@rm -rf ${GOPATH}/src/github.com/j75689/gqlgen
	@git clone https://github.com/j75689/gqlgen.git ${GOPATH}/src/github.com/j75689/gqlgen
	@cd ${GOPATH}/src/github.com/j75689/gqlgen && go install
	@go get github.com/google/wire/cmd/wire

gen:
	# generate model & reslover
	@gqlgen
	# generate dependency injection
	@wire ./...

build:
	@go build -ldflags="-X ${REPO}/cmd.version=${VERSION} -X ${REPO}/cmd.commitID=${GIT_COMMIT} -X ${REPO}/cmd.commitDate=${GIT_COMMIT_DATE}"

build-image:
	@./build/build.sh

# bench
bench-redis-mq: 
	$(GO_TEST_CMD) -run=none -bench=. -benchmem -benchtime=1s -v $(PROJECT_PATH)/pkg/mq/redis_stream/*.go