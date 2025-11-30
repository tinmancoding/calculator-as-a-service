# Calculator as a Service - Makefile
# Docker image configuration
DOCKER_REGISTRY ?= tinmancoding
TAG ?= latest
DOCKER_PLATFORM ?= linux/amd64

# Service directories
SERVICES_DIR := services
GATEWAY_DIR := $(SERVICES_DIR)/gateway
PARSER_DIR := $(SERVICES_DIR)/parser
ADDITION_DIR := $(SERVICES_DIR)/addition-service
SUBTRACTION_DIR := $(SERVICES_DIR)/subtraction-service
MULTIPLICATION_DIR := $(SERVICES_DIR)/multiplication-service
DIVISION_DIR := $(SERVICES_DIR)/division-service

# Docker image names
GATEWAY_IMAGE := $(DOCKER_REGISTRY)/calculator-gateway:$(TAG)
PARSER_IMAGE := $(DOCKER_REGISTRY)/calculator-parser:$(TAG)
ADDITION_IMAGE := $(DOCKER_REGISTRY)/calculator-addition:$(TAG)
SUBTRACTION_IMAGE := $(DOCKER_REGISTRY)/calculator-subtraction:$(TAG)
MULTIPLICATION_IMAGE := $(DOCKER_REGISTRY)/calculator-multiplication:$(TAG)
DIVISION_IMAGE := $(DOCKER_REGISTRY)/calculator-division:$(TAG)

.PHONY: docker-all docker-gateway docker-parser docker-addition docker-subtraction docker-multiplication docker-division docker-push-all docker-push-gateway docker-push-parser docker-push-addition docker-push-subtraction docker-push-multiplication docker-push-division

# Build all Docker images
docker-all: docker-gateway docker-parser docker-addition docker-subtraction docker-multiplication docker-division
	@echo "All Docker images built successfully!"

# Build individual service Docker images
docker-gateway:
	docker build --platform=$(DOCKER_PLATFORM) -t $(GATEWAY_IMAGE) $(GATEWAY_DIR)

docker-parser:
	docker build --platform=$(DOCKER_PLATFORM) -t $(PARSER_IMAGE) $(PARSER_DIR)

docker-addition:
	docker build --platform=$(DOCKER_PLATFORM) -t $(ADDITION_IMAGE) $(ADDITION_DIR)

docker-subtraction:
	docker build --platform=$(DOCKER_PLATFORM) -t $(SUBTRACTION_IMAGE) $(SUBTRACTION_DIR)

docker-multiplication:
	docker build --platform=$(DOCKER_PLATFORM) -t $(MULTIPLICATION_IMAGE) $(MULTIPLICATION_DIR)

docker-division:
	docker build --platform=$(DOCKER_PLATFORM) -t $(DIVISION_IMAGE) $(DIVISION_DIR)

# Push all Docker images
docker-push-all: docker-push-gateway docker-push-parser docker-push-addition docker-push-subtraction docker-push-multiplication docker-push-division
	@echo "All Docker images pushed successfully!"

# Push individual service Docker images
docker-push-gateway:
	docker push $(GATEWAY_IMAGE)

docker-push-parser:
	docker push $(PARSER_IMAGE)

docker-push-addition:
	docker push $(ADDITION_IMAGE)

docker-push-subtraction:
	docker push $(SUBTRACTION_IMAGE)

docker-push-multiplication:
	docker push $(MULTIPLICATION_IMAGE)

docker-push-division:
	docker push $(DIVISION_IMAGE)

