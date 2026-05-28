VERSION   ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILDDATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(BUILDDATE)

EXT := $(if $(filter windows,$(GOOS)),.exe,)

.PHONY: build build-checker build-mcp build-linux build-windows test lint clean

build: build-checker build-mcp

build-checker:
	cd apps/checker && CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o ../../dist/checker$(EXT) .

build-mcp:
	cd apps/mcp-server && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o ../../dist/mcp-server$(EXT) .

build-windows:
	$(MAKE) build GOOS=windows GOARCH=amd64

build-linux:
	$(MAKE) build GOOS=linux GOARCH=amd64

test:
	cd apps/checker && go test ./...
	cd apps/mcp-server && go test ./...
	cd pkg/result && go test ./...

lint:
	cd apps/checker && go vet ./...
	cd apps/mcp-server && go vet ./...
	cd pkg/result && go vet ./...

clean:
	rm -rf dist
