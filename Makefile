GIT_HASH := $(shell git log --format="%h" -n 1)

CALENDAR := "./bin/calendar"
SCHEDULER := "./bin/scheduler"
SENDER := "./bin/sender"

LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(CALENDAR) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/scheduler
	go build -v -o $(SENDER) -ldflags "$(LDFLAGS)" ./cmd/sender

run: build
	$(SCHEDULER) --config ./configs/scheduler_config.yaml &
	$(SENDER) --config ./configs/sender_config.yaml &
	$(CALENDAR) --config ./configs/calendar_config.yaml

version: build
	$(CALENDAR_BIN) version

test:
	go test -v -race ./internal/...

install-linter:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.46.2

lint: install-linter
	sudo golangci-lint run ./...

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

up:
	sudo docker-compose -f ./deployments/docker-compose.yaml up --build

down:
	sudo docker-compose -f ./deployments/docker-compose.yaml down

integration-tests:
	set -e ;\
	sudo docker-compose -f ./deployments/docker-compose.test.yaml up --build -d ;\
	test_status_code=0 ;\
	sudo docker-compose -f ./deployments/docker-compose.test.yaml run integration-tests go test -v || test_status_code=$$? ;\
	sudo docker-compose -f ./deployments/docker-compose.test.yaml down ;\
	exit $$test_status_code ;\

.PHONY: build run version test lint generate up down integration-tests