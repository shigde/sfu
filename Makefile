SERVER_NAME = shig
CLT_NAME = shigclt
GO_LDFLAGS = -ldflags "-s -w"
GO_VERSION = 1.20
GO_TESTPKGS:=$(shell go list ./... | grep -v cmd | grep -v examples)

export CGO_ENABLED=1

all: nodes

go_init:
	go mod download
	go generate ./...

clean:
	rm -rf bin

build: go_init
	go build -race -o ./bin/$(SERVER_NAME) ./cmd/server

run: build
	./bin/$(SERVER_NAME) -c config.toml

race:
	go run -race ./cmd/server -c config.toml

build-linux: go_init
	GOOS=linux GOARCH=amd64 go build -o bin/$(SERVER_NAME).linux.amd64 $(GO_LDFLAGS) ./cmd/server

build-ctl:
	go build -o bin/$(CLT_NAME) ./cmd/ctl

run-ctl: build-ctl
	chmod +x bin/$(CLT_NAME)
	./bin/$(CLT_NAME) -c config.toml


test: go_init
	go test \
		-timeout 240s \
		-coverprofile=cover.out -covermode=atomic \
		-v -race ${GO_TESTPKGS}

monitor:
	docker-compose -f ./mon/dev/docker-compose.yml up -d

monitor-stop:
	docker-compose -f ./mon/dev/docker-compose.yml down -v

build-streamer: go_init
	go build -o ./bin/media_streamer ./cmd/media_streamer

run-streamer: build-streamer
	 ./bin/media_streamer -c config.toml

