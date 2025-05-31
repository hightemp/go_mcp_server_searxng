BINARY_NAME=go_mcp_server_searxng
LDFLAGS=-ldflags "-s -w"
CGO_ENABLED=0

.PHONY: build clean test run

build:
	CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(BINARY_NAME) .

build-linux:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .

build-windows:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

build-darwin:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .

build-all: build-linux build-windows build-darwin

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

test:
	go test -v ./...

run:
	go run .

install:
	go install $(LDFLAGS) .

deps:
	go mod tidy
	go mod download

release: clean deps build-all