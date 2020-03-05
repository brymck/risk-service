PROTOS := brymck/securities/v1/securities_api

PROJECT_ID = $(shell gcloud config get-value project)
SERVICE_NAME := $(notdir $(CURDIR))
GO_FILES := $(shell find . -name '*.go')
PROTO_FILES := $(shell find proto -name '*.proto' 2>/dev/null) $(foreach proto,$(PROTOS),proto/$(proto).proto)
PROTO_PATH := /usr/local/include
GENPROTO_FILES := $(patsubst proto/%.proto,genproto/%.pb.go,$(PROTO_FILES))

all: proto test build

init: .init.stamp

.init.stamp:
	go get -u github.com/golang/protobuf/protoc-gen-go
	go mod download
	touch $@

proto: $(GENPROTO_FILES)

proto/brymck/securities/v1/securities_api.proto:
	mkdir -p $(dir $@)
	curl --fail --location --output $@ --silent --show-error https://raw.githubusercontent.com/brymck/securities-service/master/$@

genproto/%.pb.go: proto/%.proto | .init.stamp
	mkdir -p $(dir $@)
	protoc -Iproto -I$(PROTO_PATH) --go_out=plugins=grpc:genproto $<

test: profile.out

profile.out: $(GO_FILES) $(GENPROTO_FILES) | .init.stamp
	go test -race -coverprofile=profile.out -covermode=atomic ./...

build: service

service: $(GO_FILES) $(GENPROTO_FILES) | .init.stamp
	go build -ldflags='-w -s' -o service cmd/web/*.go

run: service
	./service

client: $(GO_FILES) $(GENPROTO_FILES) | .init.stamp
	go build -ldflags='-w -s' -o client cmd/client/*.go

docker:
	docker build . --tag gcr.io/$(PROJECT_ID)/$(SERVICE_NAME)

clean:
	rm -rf proto/alpha_vantage genproto/ .init.stamp profile.out client service

.PHONY: all init proto test build run docker clean
