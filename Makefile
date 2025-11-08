.PHONY: build test coverage lint clean install dev build-all fmt help

# 変数
BINARY_NAME=tig-gh
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"
GO=go
GOFLAGS=-v

# デフォルトターゲット
.DEFAULT_GOAL := help

## help: ヘルプを表示
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: バイナリをビルド
build:
	$(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME) cmd/tig-gh/main.go

## install: バイナリをインストール
install:
	$(GO) install $(LDFLAGS) ./cmd/tig-gh

## test: テストを実行
test:
	$(GO) test -v -short ./...

## test-all: 統合テストを含むすべてのテストを実行
test-all:
	$(GO) test -v ./...

## coverage: カバレッジを生成
coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: コードをlint
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

## fmt: コードをフォーマット
fmt:
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || echo "goimports not found, skipping"

## clean: ビルド成果物をクリーン
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

## dev: 開発モード（ホットリロード）
dev:
	@which air > /dev/null || (echo "air not found. Install: go install github.com/air-verse/air@latest" && exit 1)
	air

## build-all: すべてのプラットフォーム向けにビルド
build-all:
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 cmd/tig-gh/main.go
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 cmd/tig-gh/main.go
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 cmd/tig-gh/main.go
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 cmd/tig-gh/main.go
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe cmd/tig-gh/main.go

## deps: 依存関係を更新
deps:
	$(GO) mod download
	$(GO) mod tidy

## run: アプリケーションを実行
run: build
	./bin/$(BINARY_NAME)
