# Worker service

Worker service provides functionalities to start, stop and query results of Linux commands.

This project contains 3 components:
- A package to start, stop and query results of Linux commands
- An HTTPS API that uses the above package
- A CLI to interact with the API

## Setting up the API server

### Building

```bash
$ make docker-build-api
```

This command will build a Docker image for the API server called `worker-api:latest`

### Running
```bash
$ make docker-run-api
```

This command will create and start a Docker container using the above image.

The container's name is `worker-api`. The endpoints will be accessible via https://localhost:8080
