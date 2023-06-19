envfile ?= .env
-include $(envfile)
ifneq ("$(wildcard $(envfile))","")
	export $(shell sed 's/=.*//' $(envfile))
endif

.PHONY: init
init:
	@cp .env.dist .env

run:
	@go run main.go