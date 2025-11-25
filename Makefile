PROJECT ?= jetson-monitor
BASE_PATH := $(shell pwd)
BUILD_DIR ?= $(BASE_PATH)/bin
PKG := github.com/juandiii/jetson-monitor

# Allow overriding environment for cross compilation / CGO
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0
CC ?= $(shell go env CC)
CGO_CFLAGS ?=
CGO_LDFLAGS ?=

# ldflags: embed version info if available
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo v0.0.0)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS ?= -s -w \
	-X 'github.com/juandiii/jetson-monitor/version.Version=$(VERSION)' \
	-X 'github.com/juandiii/jetson-monitor/version.Commit=$(GIT_COMMIT)' \
	-X 'github.com/juandiii/jetson-monitor/version.BuildTime=$(BUILD_TIME)'

OUTPUT := $(BUILD_DIR)/$(PROJECT)-$(GOOS)-$(GOARCH)

.PHONY: all build cross build-static build-debug clean test tidy release build-matrix

all: build

# Local build (uses GOOS/GOARCH from environment unless overridden)
build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(PROJECT) for $(GOOS)/$(GOARCH) (CGO_ENABLED=$(CGO_ENABLED))"
	CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) CGO_CFLAGS='$(CGO_CFLAGS)' CGO_LDFLAGS='$(CGO_LDFLAGS)' \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(OUTPUT) -ldflags "$(LDFLAGS)" ./
	@echo "Built: $(OUTPUT)"

# Build with CGO disabled explicitly (static binary for many cases)
build-static:
	$(MAKE) CGO_ENABLED=0 build

# Build debug (no -s -w)
build-debug:
	@mkdir -p $(BUILD_DIR)
	@echo "Building debug binary"
	CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) CGO_CFLAGS='$(CGO_CFLAGS)' CGO_LDFLAGS='$(CGO_LDFLAGS)' \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(PROJECT)-debug ./

# Cross build: pass GOOS and GOARCH on the command line, e.g.
# make cross GOOS=linux GOARCH=amd64
cross:
	$(MAKE) build

# Build a matrix of common targets (linux/amd64, linux/arm64, darwin/amd64)
BUILD_MATRIX ?= linux/amd64 linux/arm64 darwin/amd64
build-matrix:
	@for t in $(BUILD_MATRIX); do \
		OS=$${t%/*}; ARCH=$${t#*/}; echo "-> building for $$OS/$$ARCH"; \
		$(MAKE) GOOS=$$OS GOARCH=$$ARCH build; \
	done

# buildx multi-arch build helper
# Set DOCKER_PUSH=true to push the resulting multi-arch manifest to a registry.
DOCKER_PUSH ?= false
comma := ,
empty :=
space := $(empty) $(empty)
PLATFORMS := $(subst $(space),$(comma),$(BUILD_MATRIX))
LOCAL_PLATFORM := $(shell go env GOOS)/$(shell go env GOARCH)

.PHONY: docker-buildx
docker-buildx:
	@echo "Ensuring buildx builder exists..."
	-docker buildx create --use --name $(PROJECT)-builder 2>/dev/null || true
	@echo "Build platforms: $(PLATFORMS)"
	@echo "Invoking helper script scripts/docker-buildx.sh"
	@DOCKER_PUSH=$(DOCKER_PUSH) PLATFORMS='$(PLATFORMS)' LOCAL_PLATFORM='$(LOCAL_PLATFORM)' \
		PROJECT='$(PROJECT)' VERSION='$(VERSION)' sh ./scripts/docker-buildx.sh

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)

# Create release tarballs for the build matrix
release: tidy build-matrix
	@echo "Creating release archives"
	@for t in $(BUILD_MATRIX); do \
		OS=$${t%/*}; ARCH=$${t#*/}; BIN=$(BUILD_DIR)/$(PROJECT)-$$OS-$$ARCH; \
		if [ -f $$BIN ]; then \
			TAR=$(BUILD_DIR)/$(PROJECT)-$$OS-$$ARCH.tar.gz; \
			tar -C $(BUILD_DIR) -czf $$TAR $$(basename $$BIN); \
			echo "Created $$TAR"; \
		fi; \
	done

# Build a local Docker image. Tags as $(PROJECT):$(VERSION)
# Passes LDFLAGS so the binary inside the image contains version metadata.
.PHONY: docker-build
docker-build:
	@echo "Building docker image $(PROJECT):$(VERSION)"
	docker build \
		--build-arg GOOS=linux \
		--build-arg GOARCH=amd64 \
		--build-arg CGO_ENABLED=$(CGO_ENABLED) \
		--build-arg LDFLAGS="$(LDFLAGS)" \
		-t $(PROJECT):$(VERSION) .

.PHONY: docker-run docker-stop
# Run the built image for local testing using the repository `config.yml`.
# Ensures a persistent `notified_state.json` file exists and mounts it.
docker-run:
	@echo "Preparing local config and state..."
	@touch notified_state.json
	@echo "Starting container $(PROJECT)-run (detached) mapping ./config.yml -> /home/jetson/config.yml"
	docker run --rm -d --name $(PROJECT)-run -p 38080:38080 \
		-v $(PWD)/config.yml:/home/jetson/config.yml:ro \
		-v $(PWD)/notified_state.json:/home/jetson/jetson_notified_state.json \
		-e LOG_LEVEL=debug \
		$(PROJECT):$(VERSION)

docker-stop:
	@echo "Stopping container $(PROJECT)-run if running"
	-docker rm -f $(PROJECT)-run || true