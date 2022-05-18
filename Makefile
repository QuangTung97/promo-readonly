.PHONY: run lint install-tools test test-race benchmark migrate-up migrate-down-1 generate

run:
	go run cmd/server/main.go start

lint:
	$(foreach f,$(shell go fmt ./...),@echo "Forget to format file: ${f}"; exit 1;)
	go vet ./...
	revive -config revive.toml -formatter friendly ./...

install-tools:
	go install github.com/matryer/moq
	go install github.com/mgechev/revive
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

test:
	go test -v -p 1 -count=1 -covermode=count -coverprofile=coverage.out ./...

test-race:
	go test -v -p 1 -count=1 -race ./...

benchmark:
	echo "TODO"

migrate-up:
	go run cmd/migrate/main.go up

migrate-down-1:
	go run cmd/migrate/main.go down 1

generate:
	./generate.sh