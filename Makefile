BIN := "./bin/calendar"
CONFIG := "./configs/config.yml"

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config $(CONFIG)

version: build
	$(BIN) version

test:
	go test -v -race ./internal/...

install-linter:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.46.2

lint: install-linter
	golangci-lint run ./...

generate:
	mkdir -p ./google/api
	(test -f ./google/api/annotations.proto) || curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > "./google/api/annotations.proto"
	(test -f ./google/api/http.proto) || curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > "./google/api/http.proto"

	protoc -I=. \
		--go_out ./internal/server/grpc --go_opt=paths=source_relative \
		--go-grpc_out ./internal/server/grpc --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=./internal/server/grpc/ \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt generate_unbound_methods=true \
		./google/EventService.proto

.PHONY: build run version test lint generate