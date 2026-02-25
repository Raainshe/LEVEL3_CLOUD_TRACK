REGISTRY ?= registry.onstackit.cloud
IMAGE_NAME ?= ryan-paas/paas-frontend
IMAGE_TAG ?= latest
IMAGE := $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

run:
	@echo "Running frontend dev server..."
	npm run dev

podman-build:
	@echo "Building linux/amd64 frontend image with Podman..."
	podman build --arch amd64 --os linux -t $(IMAGE) -f Containerfile .

podman-push: podman-build
	@echo "Pushing frontend image to Stackit registry..."
	podman push $(IMAGE)

.PHONY: run podman-build podman-push

