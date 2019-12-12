# generate go structures from proto files
define generate_sources
	$(if $(shell command -v protoc --go_out=. 2> /dev/null),$(info golang/protobuf is installed),$(error golang/protobuf missing, please install it with go get github.com/golang/protobuf))
	@echo "# generating sources"
	@cd model && go generate -x ./...
endef

all: generate build

generate:
	$(call generate_sources)
	@echo "# done"

clean:
	cd editor && go clean
	cd view && go clean
	cd watcher && go clean
	cd election && go clean

build:
	cd view && go build
	cd editor && go build
	cd watcher && go build
	cd election && go build
