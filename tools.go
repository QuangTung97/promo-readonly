// +build tools

package tools

import (
	_ "github.com/QuangTung97/otelwrap"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/matryer/moq"
	_ "github.com/mgechev/revive"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
