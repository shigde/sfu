SERVER_NAME = shig
CLT_NAME = shigclt
GO_LDFLAGS = -ldflags "-s -w"
GO_VERSION = 1.21
GO_TESTPKGS:=$(shell go list ./... | grep -v cmd | grep -v examples)

export CGO_ENABLED=1

all: nodes

go_init:
	go mod download
	go generate ./...

web_init:
	cd web && npm install

clean:
	rm -rf bin
	rm -rf web/dist

build_web: web_init
	cd web && npm run build

build: go_init
	go build -race -o ./bin/$(SERVER_NAME) ./cmd/server

run: build
	./bin/$(SERVER_NAME) -config=config/app.config.dev.toml

run-remote: build
	./bin/$(SERVER_NAME) -config=config/app.config.remote.toml

race:
	go run -race ./cmd/server -config=config/app.config.dev.toml

build-linux: go_init
	GOOS=linux GOARCH=amd64 go build -o bin/$(SERVER_NAME).linux.amd64 $(GO_LDFLAGS) ./cmd/server

build-clt-linux: go_init
	GOOS=linux GOARCH=amd64 go build -o bin/$(CLT_NAME).linux.amd64 $(GO_LDFLAGS) ./cmd/clt

build-clt:
	go build -o bin/$(CLT_NAME) ./cmd/clt

run-send-clt: build-clt
	chmod +x bin/$(CLT_NAME)
	./bin/$(CLT_NAME) -c .shigClt.toml send --video input.ivf --audio input.ogg --url http://localhost:8080/space/live_stream_channel@localhost:9000/stream/4cd87221-e18e-4ba9-b9e1-4b7c17a86bca

run-start-clt: build-clt
	chmod +x bin/$(CLT_NAME)
	./bin/$(CLT_NAME) -c .shigClt.toml start --rtp http://localhost:1365/.. --key 352rr245 --url http://localhost:8080/space/../stream/..

run-stop-clt: build-clt
	chmod +x bin/$(CLT_NAME)
	./bin/$(CLT_NAME) -c .shigClt.toml stop --url http://localhost:8080/space/../stream/..

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

docker-build-linux:
	DOCKER_DEFAULT_PLATFORM=linux/amd64 docker run --rm -v ${CURDIR}:/usr/src/shigde -w /usr/src/shigde golang:1.21 make build-linux

docker-build-container: docker-build-linux
	DOCKER_DEFAULT_PLATFORM=linux/amd64 docker image build . --tag shigde/instance

docker-run:
	docker run --rm docker.io/shigde/instance

build-container: build-linux
	docker image build . --tag shigde/instance \
