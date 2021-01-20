DOCKER_IMAGE = worker-api
DOCKER_CONTAINER_NAME = worker-api
BUILDDIR = build

# 'make all' builds a binary for the CLI, builds and starts a Docker container for the API server
.PHONY: all
all:
	$(MAKE) docker-build-api && \
	$(MAKE) docker-run-api && \
	$(MAKE) build-wkct

# Build a Docker image for the API server
.PHONY: docker-build-api
docker-build-api:
	docker build -t $(DOCKER_IMAGE) .

# Run a Docker container for the API server
.PHONY: docker-run-api
docker-run-api:
	docker run -dp 8080:8080 --name $(DOCKER_CONTAINER_NAME) $(DOCKER_IMAGE):latest

# Build the CLI
.PHONY: build-wkct
build-wkct:
	go build -o ${BUILDDIR}/wkct ./client
