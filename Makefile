envfile ?= .env
-include $(envfile)
ifneq ("$(wildcard $(envfile))","")
	export $(shell sed 's/=.*//' $(envfile))
endif

export GOCACHE := $(if ${GOCACHE},${GOCACHE},$(shell go env GOCACHE))
export GOPATH := $(if ${GOPATH},${GOPATH},$(shell go env GOPATH))

.PHONY: init
init:
	@cp .env.dist .env

.PHONY: dev-dependencies
dev-dependencies:
	@go install github.com/matryer/moq@latest

.PHONY: run
run:
	@go run main.go

.PHONY: generate
generate:
	@mockgen -source=internal/mux.go -destination=internal/mocks/notifier.go -package=mocks Notifier

.PHONY: test
test:
	@go test ./... 

.PHONY: up
up:
	@docker-compose up -d