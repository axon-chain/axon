#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
VERSION ?= v1.1.1

BUILD_DIR ?= $(CURDIR)/build
BINARY_NAME := axond

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=axon \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(BINARY_NAME) \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

.PHONY: all build install clean test lint proto package-axond package-agent package-all

all: build

###############################################################################
###                                Build                                    ###
###############################################################################

build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building axond..."
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@echo "  Branch:  $(BRANCH)"
	@echo "  Output:  $(BUILD_DIR)/$(BINARY_NAME)"
	@go build -mod=readonly $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/axond
	@echo "Build completed."
	@$(BUILD_DIR)/$(BINARY_NAME) version

install:
	@echo "Installing axond..."
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/axond

clean:
	@rm -rf $(BUILD_DIR)

###############################################################################
###                               Protobuf                                  ###
###############################################################################

proto:
	@echo "Generating protobuf files..."
	@buf generate proto

proto-lint:
	@buf lint proto

###############################################################################
###                                Testing                                  ###
###############################################################################

test:
	@echo "Running all tests..."
	@go test -v -count=1 ./x/agent/...

test-unit:
	@echo "Running unit tests..."
	@go test -v -count=1 -run "Test" ./x/agent/keeper/

test-cover:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out -covermode=atomic ./x/agent/keeper/
	@go tool cover -func=coverage.out | tail -1
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-economics:
	@echo "Running economics tests..."
	@go test -v -run "TestBlockReward|TestContribution|TestMaxShare|TestDeflation" ./x/agent/keeper/

test-agent:
	@echo "Running agent module tests..."
	@go test -v -run "TestDefaultParams|TestChallengePool|TestScoreResponse|TestKeyFunctions" ./x/agent/keeper/

benchmark:
	@go test -bench=. -benchmem ./x/agent/...

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@unformatted=$$(gofmt -l ./x/agent/ ./app/ ./precompiles/ ./cmd/); \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@GOCACHE=$(CURDIR)/.cache/go-build go vet ./app/... ./cmd/... ./precompiles/... ./x/...

###############################################################################
###                              Docker                                     ###
###############################################################################

docker-build:
	@docker build -t axon-chain/axon:$(VERSION) .

docker-run:
	@docker run -it --rm -p 17656:17656 -p 17657:17657 -p 18545:18545 -p 18546:18546 -p 11317:11317 -p 19090:19090 axon-chain/axon:$(VERSION)

###############################################################################
###                            Distribution                                  ###
###############################################################################

package-axond:
	@echo "Packaging axond with Dockerized Go toolchain..."
	@echo "  Builder Image: $${PACKAGING_DOCKER_IMAGE:-golang:1.25.7-trixie}"
	@bash packaging/package_axond.sh

package-agent:
	@echo "Packaging agent-daemon with Dockerized Go toolchain..."
	@echo "  Builder Image: $${PACKAGING_DOCKER_IMAGE:-golang:1.25.7-trixie}"
	@bash packaging/package_agent.sh

package-all:
	@echo "Building release matrix with Dockerized Go toolchain..."
	@echo "  Builder Image: $${PACKAGING_DOCKER_IMAGE:-golang:1.25.7-trixie}"
	@VERSION="$(VERSION)" \
	OUT_DIR="$(OUT_DIR)" \
	RELEASE_NOTES_FILE="$(RELEASE_NOTES_FILE)" \
	AXOND_CGO_ENABLED="$(AXOND_CGO_ENABLED)" \
	PACKAGING_DOCKER_IMAGE="$${PACKAGING_DOCKER_IMAGE:-golang:1.25.7-trixie}" \
	bash packaging/build_release_matrix.sh
