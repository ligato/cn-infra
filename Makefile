SHELL := /usr/bin/env bash -o pipefail

VERSION	:= $(shell git describe --always --tags --dirty)
COMMIT	:= $(shell git rev-parse HEAD)
DATE	:= $(shell date +'%Y-%m-%dT%H:%M%:z')

CNINFRA := go.ligato.io/cn-infra/v2/agent
LDFLAGS = \
	-X $(CNINFRA).BuildVersion=$(VERSION) \
	-X $(CNINFRA).CommitHash=$(COMMIT) \
	-X $(CNINFRA).BuildDate=$(DATE)

ifeq ($(V),1)
GO_BUILD_ARGS += -v
endif

GOPATH := $(shell go env GOPATH)

COVER_DIR ?= /tmp

help:
	@echo "List of make targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT = help

build: examples examples-plugin ## Build all

clean: clean-examples clean-examples-plugin ## Clean all

examples: ## Build examples
	@echo "# building examples"
	cd examples/cassandra-lib  && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/etcd-lib       && make build
	cd examples/kafka-lib      && make build
	cd examples/logs-lib       && make build
	cd examples/redis-lib      && make build
	cd examples/cryptodata-lib && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}

examples-plugin: ## Build plugin examples
	@echo "# building plugin examples"
	cd examples/configs-plugin                  && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/datasync-plugin                 && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/flags-lib                       && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/kafka-plugin/hash-partitioner   && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/kafka-plugin/manual-partitioner && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/kafka-plugin/post-init-consumer && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/logs-plugin                     && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/redis-plugin                    && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/simple-agent                    && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/statuscheck-plugin              && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/prometheus-plugin               && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/cryptodata-plugin               && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}
	cd examples/bolt-plugin                     && go build -ldflags "${LDFLAGS}" ${GO_BUILD_ARGS}

clean-examples:
	@echo "# cleaning examples"
	cd examples/cassandra-lib  && go clean
	cd examples/etcd-lib       && make clean
	cd examples/kafka-lib      && make clean
	cd examples/logs-lib       && make clean
	cd examples/redis-lib      && make clean
	cd examples/cryptodata-lib && go clean

clean-examples-plugin:
	@echo "# cleaning plugin examples"
	cd examples/configs-plugin                  && go clean
	cd examples/datasync-plugin                 && go clean
	cd examples/flags-lib                       && go clean
	cd examples/kafka-plugin/hash-partitioner   && go clean
	cd examples/kafka-plugin/manual-partitioner && go clean
	cd examples/kafka-plugin/post-init-consumer && go clean
	cd examples/logs-plugin                     && go clean
	cd examples/redis-plugin                    && go clean
	cd examples/simple-agent                    && go clean
	cd examples/statuscheck-plugin              && go clean
	cd examples/prometheus-plugin               && go clean
	cd examples/bolt-plugin                     && go clean

# -------------------------------
#  Testing
# -------------------------------

CONSUL := $(shell command -v consul 2> /dev/null)

get-consul:
	@echo "# installing consul"
	./scripts/install-consul.sh
	consul version

get-testtools:
	@echo "# installing test tools"
ifndef CONSUL
	@$(MAKE) get-consul
endif

test: get-testtools ## Test all
	@echo "# running unit tests"
	go test $(GO_BUILD_ARGS) ./...

test-cover: get-testtools
	@echo "# running coverage report"
	go test ${GO_BUILD_ARGS} -covermode=count -coverprofile=${COVER_DIR}/coverage.out ./...
	@echo "# coverage data generated into ${COVER_DIR}/coverage.out"

test-cover-html: test-cover
	go tool cover -html=${COVER_DIR}/coverage.out -o ${COVER_DIR}/coverage.html
	@echo "# coverage report generated into ${COVER_DIR}/coverage.html"
	go tool cover -html=${COVER_DIR}/coverage.out

test-examples:
	@echo "# Testing examples"
	./scripts/test_examples/test_examples.sh
	@echo "# Testing examples: reactions to disconnect/reconnect of plugins redis, cassandra ..."
	./scripts/test_examples/plugin_reconnect.sh

# -------------------------------
#  Code generation
# -------------------------------

generate: generate-proto

get-proto-generators:
	go install github.com/golang/protobuf/protoc-gen-go

generate-proto: get-proto-generators ## Generate proto
	@echo "# generating proto"
	go generate -x -run=protoc ./...

# -------------------------------
#  Dependencies
# -------------------------------

dep-install:
	@echo "# downloading project's dependencies"
	go mod download

dep-update:
	@echo "# updating all dependencies"
	@echo go mod tidy -v

dep-check:
	@echo "# checking dependencies"
	go mod verify
	go mod tidy -v
	@if ! git diff --quiet go.mod ; then \
		echo "go mod tidy check failed"; \
		exit 1 ; \
	fi

# -------------------------------
#  Linters
# -------------------------------

GOLANGCI_LINT_VERSION ?= v1.21.0

LINTER := $(shell command -v golangci-lint 2> /dev/null)

get-linter:
ifndef LINTER
	@echo "# installing GolangCI-Lint"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@golangci-lint --version
endif

lint: get-linter
	@echo "# running linter"
	./scripts/static_analysis.sh golint vet

format:
	@echo "# formatting the code"
	./scripts/gofmt.sh

MDLINKCHECK := $(shell command -v markdown-link-check 2> /dev/null)

get-linkcheck:
ifndef MDLINKCHECK
	@echo "# installing markdown link checker"
	sudo apt-get update && sudo apt-get install npm
	npm install -g markdown-link-check@3.6.2
endif

check-links: get-linkcheck
	@echo "# checking links"
	./scripts/check_links.sh

get-yamllint:
	pip install --user yamllint

yamllint: get-yamllint
	@echo "# linting the yaml files"
	yamllint -c .yamllint.yml $(shell git ls-files '*.yaml' '*.yml' | grep -v 'vendor/')


.PHONY: help \
	build clean \
	examples examples-plugin clean-examples clean-examples-plugin \
	test test-examples get-testtools get-consul \
	test-cover test-cover-html test-cover-xml \
	dep-install dep-update \
	get-linter lint format \
	get-linkcheck check-links \
	get-yamllint yamllint
