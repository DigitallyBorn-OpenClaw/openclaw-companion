APP_NAME := oc-companion
CMD_PATH := ./cmd/oc-companion
BUILD_DIR := build

GO ?= go
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/gomod

.PHONY: fmt test build build-arm64 run clean

fmt:
	$(GO) fmt ./...

test:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./...

build:
	mkdir -p $(BUILD_DIR) $(GOCACHE) $(GOMODCACHE)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

build-arm64:
	mkdir -p $(BUILD_DIR) $(GOCACHE) $(GOMODCACHE)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(CMD_PATH)

run:
	$(GO) run $(CMD_PATH)

clean:
	rm -rf $(BUILD_DIR) .cache
