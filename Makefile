envfile ?= .env
-include $(envfile)
ifneq ("$(wildcard $(envfile))","")
	export $(shell sed 's/=.*//' $(envfile))
endif

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

test:
	@go test ./... 