SHELL := /bin/zsh

APP = apifox-import
DISTDIR = dist
SRC = .
PKG = ./...
NOW = $(shell date -Iseconds)
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "-s -w \
	-X 'main.BuildTime=$(NOW)'"

default: test

.PHONY: run
run:
	@go run -race . -h

.PHONY: test
test: $(SRC)
	@gofmt -w -s $(SRC)
	@go tool goimports -w $(SRC)
	@go vet $(PKG)
	@go test $(PKG)
	@go tool staticcheck $(PKG)
	@go tool govulncheck $(PKG)

.PHONY: build
build:
	@rm -rf $(DISTDIR)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) -o ./$(DISTDIR)/$(APP)-darwin-arm64 .
	@GOOS=linux GOARCH=arm64 $(GOBUILD) -o ./$(DISTDIR)/$(APP)-linux-arm64 .
	@GOOS=linux GOARCH=amd64 $(GOBUILD) -o ./$(DISTDIR)/$(APP)-linux-amd64 .
	@GOOS=windows GOARCH=amd64 $(GOBUILD) -o ./$(DISTDIR)/$(APP)-windows-amd64.exe .
