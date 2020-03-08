PROJECT_ID = $(shell gcloud config get-value project)
SERVICE_NAME := $(notdir $(CURDIR))
GO_FILES := $(shell find . -name '*.go')

all: test build

test: profile.out

profile.out: $(GO_FILES)
	go mod download
	go test -race -coverprofile=profile.out -covermode=atomic ./...

build: service

service: $(GO_FILES)
	go mod download
	go build -ldflags='-w -s' -o service cmd/web/*.go

run: service
	./service

client: $(GO_FILES)
	go mod download
	go build -ldflags='-w -s' -o client cmd/client/*.go

docker: $(GO_FILES)
	docker build --tag gcr.io/$(PROJECT_ID)/$(SERVICE_NAME) .

clean:
	rm -rf profile.out service

.PHONY: all test build run docker clean
