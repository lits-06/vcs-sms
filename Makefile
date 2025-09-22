# ================================
# Makefile for VCS-SMS Project
# ================================

.PHONY: help build clean up down logs kafka-topics kafka-create-topics kafka-delete-topics kafka-list-topics test

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
	docker-compose build

up: ## Start all services
	@echo "$(GREEN)Starting all services...$(NC)"
	docker-compose up -d

up-rebuild: ## Build and start all services
	@echo "$(GREEN)Rebuilding and starting all services...$(NC)"
	docker-compose up --build -d

down: ## Stop all services
	@echo "$(RED)Stopping all services...$(NC)"
	docker-compose down

clean: ## Stop and remove all containers, networks, and volumes
	@echo "$(RED)Cleaning up all containers, networks, and volumes...$(NC)"
	docker-compose down -v --remove-orphans
	docker system prune -f

logs: ## View logs from all services
	docker-compose logs -f

logs-service: ## View logs from specific service (usage: make logs-service SERVICE=service_name)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: Please specify SERVICE name. Example: make logs-service SERVICE=kafka$(NC)"; \
		exit 1; \
	fi
	docker-compose logs -f $(SERVICE)

restart: ## Restart all services
	@echo "$(GREEN)Restarting all services...$(NC)"
	docker-compose restart

rebuild-service: ## Rebuild a specific service. Usage: make rebuild-service SERVICE=web
	@if [ -z "$(SERVICE)" ]; then \
		echo "Please specify SERVICE, e.g., make rebuild-service SERVICE=web"; \
	else \
		docker-compose build $(SERVICE); \
	fi

## Kafka Commands
kafka-topics: kafka-list-topics ## Alias for kafka-list-topics

kafka-list-topics: ## List all Kafka topics
	@echo "$(BLUE)Listing all Kafka topics...$(NC)"
	docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --list

kafka-create-topics: ## Create all required Kafka topics
	@echo "$(GREEN)Creating all required Kafka topics...$(NC)"
	@echo "$(YELLOW)Creating server-create topic...$(NC)"
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic server-create --partitions 3 --replication-factor 1 --if-not-exists
	@echo "$(YELLOW)Creating server-update topic...$(NC)"
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic server-update --partitions 3 --replication-factor 1 --if-not-exists
	@echo "$(YELLOW)Creating server-delete topic...$(NC)"
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic server-delete --partitions 1 --replication-factor 1 --if-not-exists
	@echo "$(YELLOW)Creating dead-letter-queue topic...$(NC)"
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic dead-letter-queue --partitions 1 --replication-factor 1 --if-not-exists
	@echo "$(GREEN)All topics created successfully!$(NC)"

kafka-create-topic: ## Create a specific Kafka topic (usage: make kafka-create-topic TOPIC=topic_name PARTITIONS=3 REPLICATION=1)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-create-topic TOPIC=my-topic$(NC)"; \
		exit 1; \
	fi
	@PARTITIONS=$${PARTITIONS:-1}; \
	REPLICATION=$${REPLICATION:-1}; \
	echo "$(YELLOW)Creating topic: $(TOPIC) with $$PARTITIONS partitions and $$REPLICATION replication factor...$(NC)"; \
	docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --topic $(TOPIC) --partitions $$PARTITIONS --replication-factor $$REPLICATION --if-not-exists

kafka-delete-topics: ## Delete all Kafka topics
	@echo "$(RED)Deleting all Kafka topics...$(NC)"
	@echo "$(YELLOW)Warning: This will delete ALL topics!$(NC)"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic server-create || true
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic server-update || true
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic server-delete || true
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic dead-letter-queue || true
	@echo "$(GREEN)All topics deleted!$(NC)"

kafka-delete-topic: ## Delete a specific Kafka topic (usage: make kafka-delete-topic TOPIC=topic_name)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-delete-topic TOPIC=my-topic$(NC)"; \
		exit 1; \
	fi
	@echo "$(RED)Deleting topic: $(TOPIC)...$(NC)"
	@read -p "Are you sure you want to delete topic $(TOPIC)? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic $(TOPIC)

kafka-describe-topics: ## Describe all Kafka topics
	@echo "$(BLUE)Describing all Kafka topics...$(NC)"
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --describe

kafka-describe-topic: ## Describe a specific Kafka topic (usage: make kafka-describe-topic TOPIC=topic_name)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-describe-topic TOPIC=server-create$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Describing topic: $(TOPIC)...$(NC)"
	@docker exec -it sms_kafka kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic $(TOPIC)

kafka-produce: ## Send messages to a Kafka topic (usage: make kafka-produce TOPIC=topic_name)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-produce TOPIC=server-create$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Producing messages to topic: $(TOPIC)$(NC)"
	@echo "$(BLUE)Type your messages (press Ctrl+C to exit):$(NC)"
	@docker exec -it sms_kafka kafka-console-producer.sh --bootstrap-server localhost:9092 --topic $(TOPIC)

kafka-consume: ## Consume messages from a Kafka topic (usage: make kafka-consume TOPIC=topic_name)
	@if [ -z "$(TOPIC)" ]; then \
		echo "$(RED)Error: Please specify TOPIC name. Example: make kafka-consume TOPIC=server-create$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Consuming messages from topic: $(TOPIC)$(NC)"
	@echo "$(BLUE)Press Ctrl+C to exit$(NC)"
	@docker exec -it sms_kafka kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic $(TOPIC) --from-beginning

kafka-shell: ## Open Kafka container shell
	@echo "$(BLUE)Opening Kafka container shell...$(NC)"
	docker exec -it sms_kafka /bin/bash

## Infrastructure Commands
postgres-shell: ## Open PostgreSQL shell
	@echo "$(BLUE)Opening PostgreSQL shell...$(NC)"
	docker exec -it sms_postgres psql -U postgres -d sms_db

redis-shell: ## Open Redis shell
	@echo "$(BLUE)Opening Redis shell...$(NC)"
	docker exec -it sms_redis redis-cli

elasticsearch-shell: ## Open Elasticsearch shell
	@echo "$(BLUE)Opening Elasticsearch container shell...$(NC)"
	docker exec -it sms_elasticsearch /bin/bash

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
