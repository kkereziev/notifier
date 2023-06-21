envfile ?= .env
-include $(envfile)
ifneq ("$(wildcard $(envfile))","")
	export $(shell sed 's/=.*//' $(envfile))
endif

export GOCACHE := $(if ${GOCACHE},${GOCACHE},$(shell go env GOCACHE))
export GOPATH := $(if ${GOPATH},${GOPATH},$(shell go env GOPATH))

# ======================================= Initialization ============================================

.PHONY: init
init:
	@cp .env.dist .env

.PHONY: dev-dependencies
dev-dependencies:
	@go install github.com/matryer/moq@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3

# ======================================= Golang ============================================

.PHONY: run
run:
	@go run main.go

.PHONY: generate
generate:
	@mockgen -source=internal/mux.go -destination=internal/mocks/notifier.go -package=mocks Notifier

.PHONY: test
test:
	@go test -v -race ./... 
	

.PHONY: test-no-cache
test-no-cache:
	go test -v -count=1 -race ./...

.PHONY: lint
lint:
	@golangci-lint run --config=.golangci.yml

# ======================================= Docker ============================================


# Example usage: make up s=server
.PHONY: up
up:
	$(eval SERVICE = ${s})
	@docker-compose up -d --remove-orphans ${SERVICE}
	@docker-compose ps


.PHONY: down
down:
	@docker-compose down --remove-orphans
	@docker-compose ps
