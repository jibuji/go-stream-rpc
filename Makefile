.PHONY: install-tools
install-tools:
	go install ./cmd/protoc-gen-stream-rpc

.PHONY: generate
generate: install-tools
	protoc --go_out=. --go_opt=paths=source_relative \
		--stream-rpc_out=. --stream-rpc_opt=paths=source_relative \
		examples/calculator/proto/service.proto

.PHONY: build
build: generate
	go build ./... 