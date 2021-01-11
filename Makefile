DOCKER_IMAGE = worker-api
DOCKER_CONTAINER_NAME = worker-api

# Build a Docker image for the API server
.PHONY: docker-build-api
docker-build-api:
	docker build -t $(DOCKER_IMAGE) .

# Run a Docker container for the API server
.PHONY: docker-run-api
docker-run-api:
	docker run -dp 8080:8080 --name $(DOCKER_CONTAINER_NAME) $(DOCKER_IMAGE):latest
