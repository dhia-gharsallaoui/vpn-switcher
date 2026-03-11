BINARY   := vpn-switcher
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS  := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"

.PHONY: build run test test-cover lint clean install

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/vpn-switcher

run: build
	sudo ./bin/$(BINARY)

test:
	go test -race -count=1 ./...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/ coverage.out coverage.html

install: build
	sudo install -m 755 bin/$(BINARY) /usr/local/bin/$(BINARY)
