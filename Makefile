VERSION	:= $(shell git describe --always --tags --dirty)
COMMIT	:= $(shell git rev-parse HEAD)
DATE	:= $(shell date +'%Y-%m-%dT%H:%M%:z')

CNINFRA_CORE := github.com/ligato/cn-infra/core
LDFLAGS = -ldflags '-X $(CNINFRA_CORE).BuildVersion=$(VERSION) -X $(CNINFRA_CORE).CommitHash=$(COMMIT) -X $(CNINFRA_CORE).BuildDate=$(DATE)'

COVER_DIR ?= /tmp/

# Build all
build: examples examples-plugin

# Clean all
clean: clean-examples clean-examples-plugin

# Build examples
examples:
	@echo "=> building examples"
	cd examples/cassandra-lib && go build
	cd examples/etcdv3-lib && make build
	cd examples/kafka-lib && make build
	cd examples/logs-lib && make build
	cd examples/redis-lib && make build

# Build plugin examples
examples-plugin:
	@echo "=> building plugin examples"
	cd examples/configs-plugin && go build -i -v ${LDFLAGS}
	cd examples/datasync-plugin && go build -i -v ${LDFLAGS}
	cd examples/flags-lib && go build -i -v ${LDFLAGS}
	cd examples/kafka-plugin/hash-partitioner && go build -i -v ${LDFLAGS}
	cd examples/kafka-plugin/manual-partitioner && go build -i -v ${LDFLAGS}
	cd examples/kafka-plugin/post-init-consumer && go build -i -v ${LDFLAGS}
	cd examples/logs-plugin && go build -i -v ${LDFLAGS}
	cd examples/redis-plugin && go build -i -v ${LDFLAGS}
	cd examples/simple-agent && go build -i -v ${LDFLAGS}
	cd examples/statuscheck-plugin && go build -i -v ${LDFLAGS}
	cd examples/prometheus-plugin && go build -i -v ${LDFLAGS}
	cd examples/wiring/grpc-server && go build -i -v ${LDFLAGS} ./...
	cd examples/wiring/grpc-with-rest && go build -i -v ${LDFLAGS}

# Clean examples
clean-examples:
	@echo "=> cleaning examples"
	cd examples/cassandra-lib && rm -f cassandra-lib
	cd examples/etcdv3-lib && make clean
	cd examples/kafka-lib && make clean
	cd examples/logs-lib && make clean
	cd examples/redis-lib && make clean

# Clean plugin examples
clean-examples-plugin:
	@echo "=> cleaning plugin examples"
	rm -f examples/configs-plugin/configs-plugin
	rm -f examples/datasync-plugin/datasync-plugin
	rm -f examples/flags-lib/flags-lib
	rm -f examples/kafka-plugin/hash-partitioner/hash-partitioner
	rm -f examples/kafka-plugin/manual-partitioner/manual-partitioner
	rm -f examples/kafka-plugin/post-init-consumer/post-init-consumer
	rm -f examples/logs-plugin/logs-plugin
	rm -f examples/redis-plugin/redis-plugin
	rm -f examples/simple-agent/simple-agent
	rm -f examples/statuscheck-plugin/statuscheck-plugin
	rm -f examples/prometheus-plugin/prometheus-plugin


# Get test tools
get-testtools:
	go get -v github.com/hashicorp/consul

# Run tests
test: get-testtools
	@echo "=> running unit tests"
	go test ./core
	go test ./datasync/syncbase
	go test ./db/keyval/consul
	go test ./db/keyval/etcdv3
	go test ./db/keyval/redis
	go test ./db/sql/cassandra
	go test ./idxmap/mem
	go test ./logging/logrus
	go test ./messaging/kafka/client
	go test ./messaging/kafka/mux
	go test ./utils/addrs
	go test ./tests/gotests/itest
	go test ./rpc/grpc
	go test ./rpc/rest

# Run script for testing examples
test-examples:
	@echo "=> Testing examples"
	./scripts/test_examples/test_examples.sh
	@echo "=> Testing examples: reactions to disconnect/reconnect of plugins redis, cassandra ..."
	./scripts/test_examples/plugin_reconnect.sh

# Get coverage report tools
get-covtools:
	go get -v github.com/wadey/gocovmerge

# Run coverage report
test-cover: get-testtools get-covtools
	@echo "=> running coverage report"
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit1.out ./core
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit2.out ./datasync/syncbase
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit3.out ./db/keyval/consul
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit4.out ./db/keyval/etcdv3
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit5.out ./db/keyval/redis
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit6.out ./db/sql/cassandra
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit7.out ./idxmap/mem
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit8.out ./logging/logrus
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit9.out ./messaging/kafka/client
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit10.out ./messaging/kafka/mux
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit11.out ./utils/addrs
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit12.out ./tests/gotests/itest
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit13.out ./rpc/grpc
	go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit14.out ./rpc/rest
	@echo "=> merging coverage results"
	gocovmerge \
			${COVER_DIR}coverage_unit1.out \
			${COVER_DIR}coverage_unit2.out \
			${COVER_DIR}coverage_unit3.out \
			${COVER_DIR}coverage_unit4.out \
			${COVER_DIR}coverage_unit5.out \
			${COVER_DIR}coverage_unit6.out \
			${COVER_DIR}coverage_unit7.out \
			${COVER_DIR}coverage_unit8.out \
			${COVER_DIR}coverage_unit9.out \
			${COVER_DIR}coverage_unit10.out \
			${COVER_DIR}coverage_unit11.out \
			${COVER_DIR}coverage_unit12.out \
			${COVER_DIR}coverage_unit13.out \
			${COVER_DIR}coverage_unit14.out \
		> ${COVER_DIR}coverage.out
	@echo "=> coverage data generated into ${COVER_DIR}coverage.out"

test-cover-html: test-cover
	go tool cover -html=${COVER_DIR}coverage.out -o ${COVER_DIR}coverage.html
	@echo "=> coverage report generated into ${COVER_DIR}coverage.html"
	go tool cover -html=${COVER_DIR}coverage.out

test-cover-xml: test-cover
	gocov convert ${COVER_DIR}coverage.out | gocov-xml > ${COVER_DIR}coverage.xml
	@echo "=> coverage report generated into ${COVER_DIR}coverage.xml"

# Get dependency manager tool
get-dep:
	go get -v github.com/golang/dep/cmd/dep

# Install the project's dependencies
dep-install: get-dep
	dep ensure

# Update the locked versions of all dependencies
dep-update: get-dep
	dep ensure -update

# Get linter tools
get-linters:
	@echo "=> installing linters"
	go get -v github.com/alecthomas/gometalinter
	gometalinter --install

# Run linters
lint: get-linters
	@echo "=> running code analysis"
	./scripts/static_analysis.sh golint vet

# Format code
format:
	@echo "=> formatting the code"
	./scripts/gofmt.sh

# Get link check tool
get-linkcheck:
	sudo apt-get install npm
	npm install -g markdown-link-check

# Validate links in markdown files
check-links: get-linkcheck
	@echo "=> checking links"
	./scripts/check_links.sh

.PHONY: build clean \
	examples examples-plugin clean-examples clean-examples-plugin test test-examples \
	get-covtools test-cover test-cover-html test-cover-xml \
	get-dep dep-install dep-update \
	get-linters lint format \
	get-linkcheck check-links
