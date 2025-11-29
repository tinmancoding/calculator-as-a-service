# Calculator as a Service - Makefile
# Docker image configuration
DOCKER_REGISTRY ?= calculator
TAG ?= latest

# Service directories
SERVICES_DIR := services
GATEWAY_DIR := $(SERVICES_DIR)/gateway
PARSER_DIR := $(SERVICES_DIR)/parser
ADDITION_DIR := $(SERVICES_DIR)/addition-service
SUBTRACTION_DIR := $(SERVICES_DIR)/subtraction-service
MULTIPLICATION_DIR := $(SERVICES_DIR)/multiplication-service
DIVISION_DIR := $(SERVICES_DIR)/division-service

# Docker image names
GATEWAY_IMAGE := $(DOCKER_REGISTRY)/gateway-service:$(TAG)
PARSER_IMAGE := $(DOCKER_REGISTRY)/parser-service:$(TAG)
ADDITION_IMAGE := $(DOCKER_REGISTRY)/addition-service:$(TAG)
SUBTRACTION_IMAGE := $(DOCKER_REGISTRY)/subtraction-service:$(TAG)
MULTIPLICATION_IMAGE := $(DOCKER_REGISTRY)/multiplication-service:$(TAG)
DIVISION_IMAGE := $(DOCKER_REGISTRY)/division-service:$(TAG)

.PHONY: docker-all docker-gateway docker-parser docker-addition docker-subtraction docker-multiplication docker-division

# Build all Docker images
docker-all: docker-gateway docker-parser docker-addition docker-subtraction docker-multiplication docker-division
	@echo "All Docker images built successfully!"

# Build individual service Docker images
docker-gateway:
	docker build -t $(GATEWAY_IMAGE) $(GATEWAY_DIR)

docker-parser:
	docker build -t $(PARSER_IMAGE) $(PARSER_DIR)

docker-addition:
	docker build -t $(ADDITION_IMAGE) $(ADDITION_DIR)

docker-subtraction:
	docker build -t $(SUBTRACTION_IMAGE) $(SUBTRACTION_DIR)

docker-multiplication:
	docker build -t $(MULTIPLICATION_IMAGE) $(MULTIPLICATION_DIR)

docker-division:
	docker build -t $(DIVISION_IMAGE) $(DIVISION_DIR)
