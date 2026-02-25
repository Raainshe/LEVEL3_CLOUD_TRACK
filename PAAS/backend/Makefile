# Simple Makefile for a Go project

REGISTRY ?= registry.onstackit.cloud
IMAGE_NAME ?= ryan-paas/paas-backend
IMAGE_TAG ?= latest
IMAGE := $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

podman:
	@echo "Running backend with Podman..."
	podman run --rm -p 8080:8080 \
		-e PORT=8080 \
		-e APP_ENV=local \
		-e MONGO_DB_ATLAS_USERNAME=$${MONGO_DB_ATLAS_USERNAME} \
		-e MONGO_DB_ATLAS_PASSWORD=$${MONGO_DB_ATLAS_PASSWORD} \
		$(IMAGE)

podman-build:
	@echo "Building linux/amd64 image with Podman..."
	podman build --arch amd64 --os linux -t $(IMAGE) -f Containerfile .

podman-push: podman-build
	@echo "Pushing image to Stackit registry..."
	podman push $(IMAGE)


.PHONY: all build run test clean watch docker-run docker-down itest podman podman-build podman-push
