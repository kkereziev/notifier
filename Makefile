envfile ?= .env
-include $(envfile)
ifneq ("$(wildcard $(envfile))","")
	export $(shell sed 's/=.*//' $(envfile))
endif

SHARED_SERVICES_PATH := docker/shared
export GOCACHE := $(if ${GOCACHE},${GOCACHE},$(shell go env GOCACHE))
export GOPATH := $(if ${GOPATH},${GOPATH},$(shell go env GOPATH))

.PHONY: init
init: dev-dependencies build
	@$(MAKE) --no-print-directory -C ${SHARED_SERVICES_PATH} init
	@cp .env.dist .env


.PHONY: dev-dependencies
dev-dependencies:
	@brew install protoc
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

.PHONY: gen-proto
gen-proto:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/proto/notification.proto

.PHONY: run-client
run-client:
	@go run cmd/playground/main.go

.PHONY: shared-%
# Proxies pattern based target call to shared services Makefile
# @Example
# 	$ make shared-up s=postgres
shared-%:
	@$(MAKE) --no-print-directory -C ${SHARED_SERVICES_PATH} $*

.PHONY: build
build:
	@docker build --target prod -f docker/Dockerfile -t notification-service:prod .
	@docker build --target dev -f docker/Dockerfile -t notification-service:dev .

prod:
	@$(MAKE) shared-up s=postgres
	@$(MAKE) shared-up s=server
	@$(MAKE) shared-up s=mail-cron
	@$(MAKE) shared-up s=sms-cron
	@$(MAKE) shared-up s=slack-cron
