# ================================
# Makefile for VCS-SMS Project
# ================================

.PHONY: help build clean up down logs kafka-topics kafka-create-topics kafka-delete-topics kafka-list-topics test restart restart-service remove-containers remove-all recreate recreate-build recreate-service rebuild-service rebuild-recreate-service stop-remove-start full-reset

# Default target
.DEFAULT_GOAL := help

# Colors for output
YELLOW := \033[33m
GREEN := \033[32m
RED := \033[31m
BLUE := \033[34m
NC := \033[0m # No Color

## Help
help: ## Display this help message
	@echo "$(BLUE)VCS-SMS Project Makefile Commands$(NC)"
	@echo "=================================="
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Docker Commands
build: ## Build all Docker containers
	@echo "$(GREEN)Building all services...$(NC)"
	docker compose build

up: ## Start all services
	@echo "$(GREEN)Starting all services...$(NC)"
	docker compose up -d

build-up: ## Build and start all services
	@echo "$(GREEN)Rebuilding and starting all services...$(NC)"
	docker compose up --build -d

down: ## Stop all services
	@echo "$(RED)Stopping all services...$(NC)"
	docker compose down

clean: ## Stop and remove all containers, networks, and volumes
	@echo "$(RED)Cleaning up all containers, networks, and volumes...$(NC)"
	docker compose down -v --remove-orphans
	docker system prune -f

logs: ## View logs from all services
	docker compose logs -f

logs-service: ## View logs from specific service (usage: make logs-service SERVICE=service_name)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: Please specify SERVICE name. Example: make logs-service SERVICE=kafka$(NC)"; \
		exit 1; \
	fi
	docker compose logs -f $(SERVICE)

restart: ## Restart all services
	@echo "$(GREEN)Restarting all services...$(NC)"
	docker compose restart

restart-service: ## Restart a specific service (usage: make restart-service SERVICE=service_name)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: Please specify SERVICE name. Example: make restart-service SERVICE=auth_service$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Restarting service: $(SERVICE)...$(NC)"
	docker compose restart $(SERVICE)

remove-containers: ## Remove all containers (keeps volumes and networks)
	@echo "$(RED)Removing all containers...$(NC)"
	docker compose rm -f

remove-all: ## Remove containers, volumes, networks, and orphans
	@echo "$(RED)Removing all containers, volumes, networks...$(NC)"
	docker compose down -v --remove-orphans

recreate: ## Remove and recreate all containers
	@echo "$(YELLOW)Removing and recreating all containers...$(NC)"
	docker compose down
	docker compose up -d

recreate-build: ## Remove, rebuild and recreate all containers
	@echo "$(YELLOW)Removing, rebuilding and recreating all containers...$(NC)"
	docker compose down
	docker compose up --build -d

recreate-service: ## Remove and recreate a specific service (usage: make recreate-service SERVICE=service_name)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: Please specify SERVICE name. Example: make recreate-service SERVICE=auth_service$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Recreating service: $(SERVICE)...$(NC)"
	docker compose rm -f $(SERVICE)
	docker compose up -d $(SERVICE)

rebuild-service: ## Rebuild a specific service (usage: make rebuild-service SERVICE=service_name)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: Please specify SERVICE name. Example: make rebuild-service SERVICE=auth_service$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Rebuilding service: $(SERVICE)...$(NC)"
	docker compose build $(SERVICE)

rebuild: ## Rebuild and recreate a specific service (usage: make rebuild-recreate-service SERVICE=service_name)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: Please specify SERVICE name. Example: make rebuild-recreate-service SERVICE=auth_service$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Rebuilding and recreating service: $(SERVICE)...$(NC)"
	docker compose stop $(SERVICE)
	docker compose rm -f $(SERVICE)
	docker compose build $(SERVICE)
	docker compose up -d $(SERVICE)

stop-remove-start: ## Stop, remove containers, and start again
	@echo "$(YELLOW)Stopping, removing containers, and starting again...$(NC)"
	docker compose stop
	docker compose rm -f
	docker compose up -d

full-reset: ## Complete reset: stop, remove everything, rebuild, and start
	@echo "$(RED)Performing full reset...$(NC)"
	docker compose down -v --remove-orphans
	docker system prune -f
	docker compose build
	docker compose up -d

## Kafka Commands
kafka-topics: kafka-list-topics ## Alias for kafka-list-topics

kafka-list-topics: ## List all Kafka topics with detailed information
	@echo "$(BLUE)Listing all Kafka topics with details...$(NC)"
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --list

kafka-create-topics: ## Create all required Kafka topics
	@echo "$(GREEN)Creating all required Kafka topics...$(NC)"
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	echo "$(YELLOW)Creating server-create topic...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --create --topic server-create --partitions 5 --replication-factor 1 --if-not-exists; \
	echo "$(YELLOW)Creating server-update topic...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --create --topic server-update --partitions 5 --replication-factor 1 --if-not-exists; \
	echo "$(YELLOW)Creating server-delete topic...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --create --topic server-delete --partitions 1 --replication-factor 1 --if-not-exists; \
	echo "$(YELLOW)Creating dead-letter-queue topic...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --create --topic dead-letter-queue --partitions 1 --replication-factor 1 --if-not-exists; \
	echo "$(GREEN)All topics created successfully!$(NC)"

kafka-create-topic: ## Create a specific Kafka topic (usage: make kafka-create-topic TOPIC=topic_name PARTITIONS=3 REPLICATION=1)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-create-topic TOPIC=my-topic$(NC)"; \
		exit 1; \
	fi
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	PARTITIONS=$${PARTITIONS:-1}; \
	REPLICATION=$${REPLICATION:-1}; \
	echo "$(YELLOW)Creating topic: $(TOPIC) with $$PARTITIONS partitions and $$REPLICATION replication factor...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --create --topic $(TOPIC) --partitions $$PARTITIONS --replication-factor $$REPLICATION --if-not-exists

kafka-delete-topics: ## Delete all Kafka topics
	@echo "$(RED)Deleting all Kafka topics...$(NC)"
	@echo "$(YELLOW)Warning: This will delete ALL topics!$(NC)"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	echo "$(YELLOW)Deleting server-create topic...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic server-create 2>&1 || true; \
	echo "$(YELLOW)Deleting server-update topic...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic server-update 2>&1 || true; \
	echo "$(YELLOW)Deleting server-delete topic...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic server-delete 2>&1 || true; \
	echo "$(YELLOW)Deleting dead-letter-queue topic...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic dead-letter-queue 2>&1 || true; \
	echo "$(GREEN)Topic deletion requested!$(NC)"; \
	echo "$(BLUE)Verifying remaining topics...$(NC)"; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --list

kafka-delete-topic: ## Delete a specific Kafka topic (usage: make kafka-delete-topic TOPIC=topic_name)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-delete-topic TOPIC=my-topic$(NC)"; \
		exit 1; \
	fi
	@echo "$(RED)Deleting topic: $(TOPIC)...$(NC)"
	@read -p "Are you sure you want to delete topic $(TOPIC)? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic $(TOPIC)

kafka-describe-topics: ## Describe all Kafka topics
	@echo "$(BLUE)Describing all Kafka topics...$(NC)"
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --describe

kafka-describe-topic: ## Describe a specific Kafka topic (usage: make kafka-describe-topic TOPIC=topic_name)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-describe-topic TOPIC=server-create$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Describing topic: $(TOPIC)...$(NC)"
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec $$KAFKA_CONTAINER kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic $(TOPIC)

kafka-consumer-group: ## List all Kafka consumer groups
	@echo "$(BLUE)Listing all Kafka consumer groups...$(NC)"
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec $$KAFKA_CONTAINER kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --all-groups

kafka-produce: ## Send messages to a Kafka topic (usage: make kafka-produce TOPIC=topic_name)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-produce TOPIC=server-create$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Producing messages to topic: $(TOPIC)$(NC)"
	@echo "$(BLUE)Type your messages (press Ctrl+C to exit):$(NC)"
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec -it $$KAFKA_CONTAINER kafka-console-producer.sh --bootstrap-server localhost:9092 --topic $(TOPIC)

kafka-consume: ## Consume messages from a Kafka topic (usage: make kafka-consume TOPIC=topic_name)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-consume TOPIC=server-create$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Consuming messages from topic: $(TOPIC)$(NC)"
	@echo "$(BLUE)Press Ctrl+C to exit$(NC)"
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec -it $$KAFKA_CONTAINER kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic $(TOPIC) --from-beginning

kafka-shell: ## Open Kafka container shell
	@echo "$(BLUE)Opening Kafka container shell...$(NC)"
	@KAFKA_CONTAINER=$$(docker ps --filter "name=sms_stack_kafka" --format "{{.ID}}" | head -1); \
	if [ -z "$$KAFKA_CONTAINER" ]; then \
		echo "$(RED)Error: Kafka container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec -it $$KAFKA_CONTAINER /bin/bash

## Infrastructure Commands
postgres-shell: ## Open PostgreSQL shell
	@echo "$(BLUE)Opening PostgreSQL shell...$(NC)"
	@POSTGRES_CONTAINER=$$(docker ps --filter "name=sms_stack_postgres" --format "{{.ID}}" | head -1); \
	if [ -z "$$POSTGRES_CONTAINER" ]; then \
		echo "$(RED)Error: PostgreSQL container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec -it $$POSTGRES_CONTAINER psql -U dev_user -d sms_db

redis-shell: ## Open Redis shell
	@echo "$(BLUE)Opening Redis shell...$(NC)"
	@REDIS_CONTAINER=$$(docker ps --filter "name=sms_stack_redis" --format "{{.ID}}" | head -1); \
	if [ -z "$$REDIS_CONTAINER" ]; then \
		echo "$(RED)Error: Redis container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec -it $$REDIS_CONTAINER redis-cli

elasticsearch-shell: ## Open Elasticsearch shell
	@echo "$(BLUE)Opening Elasticsearch container shell...$(NC)"
	@ES_CONTAINER=$$(docker ps --filter "name=sms_stack_elasticsearch" --format "{{.ID}}" | head -1); \
	if [ -z "$$ES_CONTAINER" ]; then \
		echo "$(RED)Error: Elasticsearch container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec -it $$ES_CONTAINER /bin/bash

ELASTIC_URL=http://localhost:9200

SNAPSHOT_INDEX=snapshotidx
RECORD_INDEX=recordidx

elasticsearch-create-indices: ## Create Elasticsearch indices
	@echo "$(GREEN)Creating Elasticsearch indices...$(NC)";
	@if ! curl --max-time 5 -s -o /dev/null -w '%{http_code}' -X HEAD '$(ELASTIC_URL)/$(SNAPSHOT_INDEX)' | grep -q 200; then \
		if ! printf '{"settings":{"number_of_shards":1,"number_of_replicas":0},"mappings":{"properties":{"server_id":{"type":"keyword"},"port":{"type":"integer"},"status":{"type":"keyword"},"timestamp":{"type":"date"}}}}' | \
			curl -s -S -X PUT '$(ELASTIC_URL)/$(SNAPSHOT_INDEX)' -H 'Content-Type: application/json' -d @-; then \
			echo '$(RED)Failed to create index $(SNAPSHOT_INDEX)!$(NC)'; \
		fi; \
	fi
	@if ! curl --max-time 5 -s -o /dev/null -w '%{http_code}' -X HEAD '$(ELASTIC_URL)/$(RECORD_INDEX)' | grep -q 200; then \
  		if ! printf '{"settings":{"number_of_shards":1,"number_of_replicas":0},"mappings":{"properties":{"server_id":{"type":"keyword"},"port":{"type":"integer"},"status":{"type":"keyword"},"timestamp":{"type":"date"}}}}' | \
			curl -s -S -X PUT '$(ELASTIC_URL)/$(RECORD_INDEX)' -H 'Content-Type: application/json' -d @-; then \
			echo '$(RED)Failed to create index $(RECORD_INDEX)!$(NC)'; \
		fi; \
	fi
	@echo "$(GREEN)Elasticsearch indices created successfully!$(NC)"

## Test Commands
test: ## Run tests for all services
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

## Development Commands
lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run

format: ## Format Go code
	@echo "$(GREEN)Formatting Go code...$(NC)"
	go fmt ./...

tidy: ## Tidy Go modules
	@echo "$(GREEN)Tidying Go modules...$(NC)"
	go mod tidy

## Quick Setup Commands
setup: build up kafka-create-topics ## Build, start services, and create Kafka topics
	@echo "$(GREEN)✅ VCS-SMS setup complete!$(NC)"
	@echo "$(BLUE)Services available at:$(NC)"
	@echo "  - Traefik Dashboard: http://traefik.localhost:8080"
	@echo "  - Jaeger UI: http://jaeger.localhost"
	@echo "  - User Service: http://user.localhost/api/users"
	@echo "  - Auth Service: http://auth.localhost/api/auth"
	@echo "  - Server Service: http://server.localhost/api/servers"
	@echo "  - Report Service: http://report.localhost/api/reports"

teardown: kafka-delete-topics down clean ## Delete topics, stop services, and cleanup
	@echo "$(RED)✅ VCS-SMS teardown complete!$(NC)"

status: ## Show status of all services
	@echo "$(BLUE)Service Status:$(NC)"
	@docker-compose ps

init: kafka-create-topics elasticsearch-create-indices ## Initialize Kafka topics and Elasticsearch indices
	@echo "$(GREEN)✅ Initialization complete!$(NC)"

clear-snapshotidx: ## Clear all documents in snapshotidx
	curl -X POST "$(ELASTIC_URL)/$(SNAPSHOT_INDEX)/_delete_by_query" \
	-H 'Content-Type: application/json' \
	-d '{"query": {"match_all": {}}}'

clear-snapshot-redis: ## Clear all keys snapshotKey:* in Redis
	@REDIS_CONTAINER=$$(docker ps --filter "name=sms_stack_redis" --format "{{.ID}}" | head -1); \
	if [ -z "$$REDIS_CONTAINER" ]; then \
		echo "$(RED)Error: Redis container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec -i $$REDIS_CONTAINER redis-cli --scan --pattern "snapshotKey:*" | xargs -r docker exec -i $$REDIS_CONTAINER redis-cli del

clear-postgres-servers: ## Clear all servers from PostgreSQL
	@POSTGRES_CONTAINER=$$(docker ps --filter "name=sms_stack_postgres" --format "{{.ID}}" | head -1); \
	if [ -z "$$POSTGRES_CONTAINER" ]; then \
		echo "$(RED)Error: PostgreSQL container not found$(NC)"; \
		exit 1; \
	fi; \
	docker exec -i $$POSTGRES_CONTAINER psql -U dev_user -d sms_db -c "TRUNCATE TABLE servers CASCADE;"

clear-all-servers: clear-snapshotidx clear-snapshot-redis clear-postgres-servers

list-ports:
	lsof -i -P | awk 'NR>1 {split($$9,a,":");print a[length(a)]}' | sort -n | uniq

init-swarm: ## Initialize Docker Swarm (if not already initialized)
	@docker info | grep "Swarm: active" > /dev/null || docker swarm init

deploy: init-swarm ## Deploy stack using Docker Swarm
	docker stack deploy -c docker-compose.yml sms_stack

stop:
	docker swarm leave --force

build-images: ## Build Docker images for all services
	docker build -t vcs-sms-user_service:latest -f ./user_service/Dockerfile .
	docker build -t vcs-sms-auth_service:latest -f ./auth_service/Dockerfile .
	docker build -t vcs-sms-server_service:latest -f ./server_service/Dockerfile .
	docker build -t vcs-sms-report_service:latest -f ./report_service/Dockerfile .
	docker build -t vcs-sms-healthcheck_service:latest -f ./healthcheck_service/Dockerfile .