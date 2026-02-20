.PHONY: all build run clean test deps frontend-build build-full

APP_NAME := rubick
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

all: build

build:
	CGO_ENABLED=1 go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/server

run:
	CGO_ENABLED=1 go run ./cmd/server

clean:
	rm -rf bin/
	go clean

test:
	go test -v ./...

deps:
	go mod download
	go mod tidy

# 前端构建
frontend-build:
	cd web && npm install && npm run build
	rm -rf internal/static/dist
	cp -r web/dist internal/static/

# 完整构建（包含前端）
build-full: frontend-build build
