GO_CACHE ?= $(CURDIR)/.build/go-cache
LIB_DIR := $(CURDIR)/.build/lib
XF86VM_LIB := $(shell ldconfig -p 2>/dev/null | awk '/libXxf86vm\.so\.1/{print $$NF; exit}')
GO_ENV := GOCACHE=$(GO_CACHE) CGO_LDFLAGS=-L$(LIB_DIR)

.PHONY: prepare run build test test-ui vet clean

prepare:
	@mkdir -p $(GO_CACHE) $(LIB_DIR)
	@if [ -n "$(XF86VM_LIB)" ]; then ln -sf "$(XF86VM_LIB)" "$(LIB_DIR)/libXxf86vm.so"; fi

run: prepare
	$(GO_ENV) go run -buildvcs=false .

build: prepare
	$(GO_ENV) go build -buildvcs=false -o .build/ashen-crown .

test:
	GOCACHE=$(GO_CACHE) go test ./internal/core

test-ui: prepare
	$(GO_ENV) GOGAME_SCREENSHOTS=$(CURDIR)/.build/screenshots go test -v -count=1 .

vet:
	GOCACHE=$(GO_CACHE) go vet ./...

clean:
	rm -rf .build
