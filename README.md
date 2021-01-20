# Worker service

Worker service provides functionalities to start, stop and query results of Linux commands.

This project contains 3 components:
- A package to start, stop and query results of Linux commands
- An HTTPS API that uses the above package
- A CLI to interact with the API

## Building the API server and the CLI

```bash
$ make all
```

This command will create a Docker container for the API server and a binary for the CLI.

The Docker container name is `worker-api`. The endpoints will be accessible via https://localhost:8080.

The CLI binary is called `wkct`. It will be placed in the `build/` directory.

## Using the CLI

### Configurations

The CLI configuration requires 3 environment variables:
- `WORKER_USERNAME`: the username to authenticate with the API. There are 2 test usernames: `user1` and `user2`
- `WORKER_PASSWORD`: the password to authenticate with the API. The passwords are `thisispasswordforuser1` and `thisispasswordforuser2` for `user1` and `user2`, respectively
- `WORKER_CERT`: the path to the API server's certificate. This will be `certs/server.crt` if CLI is run from the root directory of the project

### Usage

The examples below are run from the root directory of the project. Hence, `./build/` is appended to the command name.

#### Starting a job to run a command

```bash
# Running a Linux command without flags or arguments
./build/wkct start ls

# Running a Linux command with flags
./build/wkct start "ls -l"

# Running a Linux command with arguments
./build/wkct start "echo hello"
```

#### Stopping a job

```bash
./build/wkct stop [job_id]
```

The `job_id` above is the ID returned after starting a job.

#### Get job results

```bash
./build/wkct job [job_id]
```

The `job_id` above is the ID returned after starting a job.

## Running tests

Run the following from the root directory of the project:

```bash
go test ./...
```

As of this moment, there are only integration tests that cover the Worker API and the Worker library.
