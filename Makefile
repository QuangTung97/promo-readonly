.PHONY: run lint install-tools

run:
	go run cmd/main.go

lint:
	$(foreach f,$(shell go fmt ./...),@echo "Forget to format file: ${f}"; exit 1;)
	go vet ./...
	revive -config revive.toml -formatter friendly ./...

install-tools:
	go install github.com/matryer/moq
	go install github.com/mgechev/revive
