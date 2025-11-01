# simply_syslog

A dead simple open source syslog server written in Go.

The server is intended to be used with docker to provide a syslog server
that is easy to configure, deploy, and scale.

> **Note:** This project is currently in active development on the `go-migration` branch. The project was originally written in Python and is being rewritten in Go for better performance and maintainability. The Go version is functional but still in alpha. See [NEXT_STEPS.md](NEXT_STEPS.md) for development roadmap.

<!-- TOC -->
- [simply\_syslog](#simply_syslog)
  - [Features:](#features)
  - [Upcoming features:](#upcoming-features)
- [How to Use:](#how-to-use)
  - [Image tags](#image-tags)
  - [Quick Start Guide (Docker)](#quick-start-guide-docker)
  - [Quick Start Guide (Bare-Metal)](#quick-start-guide-bare-metal)
  - [Docker Requirements and Usage](#docker-requirements-and-usage)
    - [Building the image](#building-the-image)
    - [Deploying the image](#deploying-the-image)
  - [Config settings guide:](#config-settings-guide)
    - [Some setting details:](#some-setting-details)
  - [Development and Contributing:](#development-and-contributing)
  - [Reporting Vulnerabilities:](#reporting-vulnerabilities)
<!-- TOC -->


## Features:

- UDP and TCP syslog support
- Docker-first deployment and configuration
- Highly configurable via environment variables or config file
- Efficient message buffering and file logging
- Designed for high-throughput and reliability
- Support for RFC3164, RFC5424, and raw syslog formats


## Upcoming features:

- Comprehensive unit tests (high priority)
- Improved graceful shutdown (in progress)
- Database logging (planned)
- TLS/encryption support for TCP (planned)
- Performance benchmarking and optimization (planned)


# How to Use:

## Image tags
- The `latest` tag is a rolling release, tracking the main branch. It is generally stable but may include features in progress.
- Versioned tags (e.g., `v0.6.0`) are stable releases with all features of that release.


## Quick Start Guide (Docker)

This is the recommended method for setting up the server.

To Deploy a very simple configuration with the default values, this command can be used:

Pull the image:

```
docker pull ryansteffan/simply_syslog
```

And then run the container:

```
docker run -d --name simply_syslog --restart always \
  -p 514:514/udp \
  -v /var/lib/docker/volume/simply_syslog:/var/log/ \
  -e DEBUG_MESSAGES=True \
  ryansteffan/simply_syslog
```

All configuration is done via environment variables (see below for options).

**Note:** TCP support is in progress. For now, only UDP is fully supported.


## Quick Start Guide (Bare-Metal)

Bare metal is supported, but Docker is recommended for most users.

1. Install Go 1.25+ on your system.
2. Clone the repository:
  ```
  git clone https://github.com/TheTurnnip/simply_syslog.git
  cd simply_syslog
  ```
3. Build the server:
  ```
  go build -o simply-syslog ./cmd/simplysyslog/main.go
  ```
4. Edit the config file at `config/config.json` or set environment variables as needed.
5. Run the server:
  ```
  ./simply-syslog
  ```


## Docker Requirements and Usage

### Building the image

To build the image locally:

1. Clone the repo:
  ```
  git clone https://github.com/TheTurnnip/simply_syslog.git
  cd simply_syslog
  ```
2. Build the image:
  ```
  docker build --tag simply_syslog .
  ```

### Deploying the image

You must map the ports you want to use and bind a volume for log persistence. All configuration is via environment variables.

Refer to the [Quick Start Guide (Docker)](#quick-start-guide-docker) for details.

## Config settings guide:

When editing the config for the docker container, all of these settings can be edited by passing
environment variables to the container when it is created.

When editing the config on bare-metal you can edit the config file directly. It can be found at
*dir_you_copied_repo_to*/simply_syslog/config/config.json

This is a quick reference guide to the configurations:

| Setting             | Explanation                                                                                               | Currently in use?                   | Allowed Values                                | Unit    |
|---------------------|-----------------------------------------------------------------------------------------------------------|-------------------------------------|-----------------------------------------------|---------|
| PROTOCOL            | The type of socket the server should listen on.                                                           | UDP (TCP in progress)               | UDP                                           | N/A     |
| BIND_ADDRESS        | The address to have the server listen on. 0.0.0.0 will listen on all interfaces.                          | Yes                                 | Must be an IPv4 address                       | N/A     |
| UDP_PORT            | The UDP port that the server will listen on.                                                              | Yes                                 | Must be an integer value                      | N/A     |
| TCP_PORT            | The TCP port that the server will listen on.                                                              | In progress                         | Must be an integer value                      | N/A     |
| MAX_TCP_CONNECTIONS | The max number of tcp connections that can be queued by the server.                                       | In progress                         | Must be an integer value                      | N/A     |
| BUFFER_LENGTH       | The max number of messages that can be stored in the buffer, before it writes to disk.                    | Yes                                 | Must be an integer value                      | N/A     |
| BUFFER_LIFESPAN     | The amount of time from the last message being received that the buffer will wait before writing to disk. | Yes                                 | Must be an integer value                      | Seconds |
| MAX_MESSAGE_SIZE    | The max size message that the server will accept.                                                         | Yes                                 | Must be an integer value                      | Bytes   |
| SYSLOG_PATH         | The path to where the syslog.log file used by the server can be found.                                    | Yes                                 | Must be a valid path that is created already. | N/A     |
| DEBUG_MESSAGES      | If debug messages should be displayed to the terminal.                                                    | Yes                                 | True or False                                 | N/A     |

### Some setting details:

- The sever expects that the file it needs to log to is made already. Ensure that not only the directory
  specified not only exists, but has a file called syslog.log in it.
- When changing the BUFFER_LIFESPAN value, be careful. If it is set to high it can lead to the buffer holding syslog
  messages for a long time. This means they will not display in the file, and also it is dangerous in the event of power
  loss and can lead to loss of data. A good rule of thumb is to keep the buffer timeout as low as possible.
  I would not set it higher than 5 seconds max, unless there is some special case where it is needed.
- The BUFFER_LENGTH needs to be balanced, while a larger buffer will allow for more requests to be taken in
  before a disk write is done, this leads to the trade-off of having a longer disk write that needs to be done later.
  I would recommend not set values any higher than 256, unless there are lots of machines on the network, or you have
  a use case that demands it.

## Development and Contributing:

We welcome contributions! This project is actively being developed and there are many opportunities to contribute.

**📋 See [NEXT_STEPS.md](NEXT_STEPS.md)** for a detailed list of prioritized tasks and features that need work.

**👥 See [CONTRIBUTING.md](CONTRIBUTING.md)** for guidelines on how to contribute, set up your development environment, and submit pull requests.

### Quick Development Setup

```bash
# Clone the repository
git clone https://github.com/ryansteffan/simply_syslog.git
cd simply_syslog

# Build the application
go build -o build/simply-syslog ./cmd/simplysyslog/main.go

# Run the application
./build/simply-syslog

# Or use Task (if installed)
task run
```

## Reporting Vulnerabilities:


If you find any vulnerabilities with this code please report the issue using the security section of the
GitHub repo (https://github.com/TheTurnnip/simply_syslog).

---

**Note:** This project is under active development. For the latest status and features, see the repository and open issues.

