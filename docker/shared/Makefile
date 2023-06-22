-include .env

export COMPOSE_PROJECT_NAME := notifications_service
export HOST_IP := $(if ${OC_HOST_IP},${OC_HOST_IP},$(shell ifconfig `netstat -rn | grep -E "^default|^0.0.0.0" | head -1 | awk '{print $$NF}'` | grep 'inet ' | awk '{print $$2}' | grep -Eo '([0-9]*\.){3}[0-9]*'))
export DB_SCHEMA_FILE := ../../../db_schema.sql

.PHONY: init
init:
	@cp .env.dist .env

.PHONY: up
up:
	$(eval SERVICE = ${s})
	@docker-compose up -d --no-build --remove-orphans ${SERVICE}
	@docker-compose ps

.PHONY: down
down:
	@docker-compose down --remove-orphans
	@docker-compose ps

# Drops all containers in the shared scope, and removes the volumes.
# Super useful for when you need to completely flush the containers and re-build them.
.PHONY: down-volumes
down-volumes:
	@docker-compose down -v --remove-orphans
	@docker-compose ps

.PHONY: start
start:
	$(eval SERVICE = ${s})
	@docker-compose start ${SERVICE}
	@docker-compose ps

.PHONY: stop
stop:
	$(eval SERVICE = ${s})
	@docker-compose stop ${SERVICE}
	@docker-compose ps

.PHONY: restart
restart:
	$(eval SERVICE = ${s})
	@docker-compose stop ${SERVICE}
	@docker-compose start ${SERVICE}
	@docker-compose ps

.PHONY: status
status:
	@docker-compose ps

.PHONY: logs
logs:
	$(eval SERVICE = ${s})
	$(eval LINES = $(if ${l},${l},50))
	@docker-compose logs --follow --tail=${LINES} ${SERVICE}

.PHONY: bash
bash:
	$(eval SERVICE = ${s})
	@docker-compose run ${SERVICE} bash
