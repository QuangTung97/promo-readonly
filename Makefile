.PHONY: run lint install-tools test test-race benchmark

run:
	go run cmd/server/main.go start

lint:
	$(foreach f,$(shell go fmt ./...),@echo "Forget to format file: ${f}"; exit 1;)
	go vet ./...
	revive -config revive.toml -formatter friendly ./...

install-tools:
	go install github.com/matryer/moq
	go install github.com/mgechev/revive

test:
	go test -v -p 1 -count=1 -covermode=count -coverprofile=coverage.out ./...

test-race:
	go test -v -p 1 -count=1 -race ./...

benchmark:
	echo "TODO"
