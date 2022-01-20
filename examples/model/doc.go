// Package etcdexample explains how to generate Golang structures from
// protobuf-formatted data.
package etcdexample

//go:generate protoc --proto_path=. --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. example.proto
