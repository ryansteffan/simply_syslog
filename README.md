# simply_syslog

A dead simple open source syslog server written in Python.

The server is intended to be used with docker to provide a syslog server
that is easy to configure, deploy, and scale.

<!-- TOC -->
* [simply_syslog](#simply_syslog)
  * [Features:](#features)
  * [Upcoming features:](#upcoming-features)
* [How to Use:](#how-to-use)
  * [Quick Start Guide (Docker):](#quick-start-guide-docker)
  * [Quick Start Guide (Bare-Metal):](#quick-start-guide-bare-metal)
  * [Docker Requirements and Usage:](#docker-requirements-and-usage)
    * [Building the image:](#building-the-image)
    * [Deploying the image:](#deploying-the-image)
  * [Config settings guide:](#config-settings-guide)
    * [Some setting details:](#some-setting-details)
  * [Reporting Vulnerabilities:](#reporting-vulnerabilities)
<!-- TOC -->

## Features:

- UDP syslog support
- Supports docker
- Highly configurable
- Able to handle large numbers of requests

## Upcoming features:

- Migration to Go
- Support for TCP
- Support for encryption

# How to Use:

## An explination of the image tags:
- The "latest" tag is treated as a rolling release, anything merged into the main branch is in the latest tag.
  Generally it is pretty stable but it you may encounter issues where certain features do not fully function.
- Tags that are versioned (Ex. v0.5.0) are stable releases with all the current features of that release working.

## Quick Start Guide (Docker):

This is the recommended method for setting up the server.

To Deploy a very simple configuration with the default values, this command can be used:

Pull the image:

```
docker pull ryansteffan/simply_syslog
```

And then run the container:

```
docker run -d --name simply_syslog --restart always -p 514:514/udp -v /var/lib/docker/volume/simply_syslog:/var/log/simply_syslog/ ryansteffan/simply_syslog
```

All other configuration of the server is done by passing environment variables to the container with the ``-e`` flag.
Such as, ``-e DEBUG_MESSAGES=False``, which would disable the debug logging for the server.

A full command with that example would look like:

```
docker run -d --name simply_syslog --restart always -p 514:514/udp -e DEBUG_MESSAGES=False -v /var/lib/docker/volume/simply_syslog:/var/log/simply_syslog/ ryansteffan/simply_syslog
```

For a full list of the settings that can be configured, refer to [this part of the
README](#config-settings-guide).

## Quick Start Guide (Bare-Metal):

While docker is the main focus of simply_syslog, bare-metal is also an option for deploying
the server.

Do note though, in the current state the server does not ship with an installation
script to make it use systemd or any init system. Making it run in the background
will need to be setup by you.

To get started running on bare-metal:

1. Make sure that you have python3 installed on your server. The sever was built for python 3.12, though it may work
   older 3.x versions.

   :warning: Issues that arise from using any version of python older than 3.12 will not be addressed.:warning:
2. cd into the directory that you will store the git repo in.
3. Run ``git clone https://github.com/TheTurnnip/simply_syslog.git`` to clone the repo.
4. cd into the server directory.
5. To make changes to the configuration, you can use your preferred text editor.
   [Refer to this part of the README](#config-settings-guide) for details on the configuration.
6. To run the server use ``python ./main.py`` on windows or on linux/mac ``python3 ./main.py``

## Docker Requirements and Usage:

Below are details on how to use both the deploy the image and how to build it.

### Building the image:

To Build the image use the `docker build` command.

Here are the steps:

1. Download the repo with `git clone https://github.com/TheTurnnip/simply_syslog.git`
2. cd into the repo. `cd git clone ./simply_syslog.git`
3. Build the image using: `docker build --tag simply_syslog .`

### Deploying the image:

Here is what the docker container will need in a deployment:

- It will need you to map the ports that you are using for the server.
- You will need to bind a volume that will be used to store the log file.
- Unless you are doing some more advanced networking with the docker container,
  it is recommended that you do not pass an environment variable to change the server port.
  The host port can be altered to your needs.

Refer to the [Quick Start Guide (Docker)](#quick-start-guide-docker) for details on what that might look like.

Also refer to the [docker documentation](https://docs.docker.com/reference/cli/docker/container/run/) for details on
deploying containers.

## Config settings guide:

When editing the config for the docker container, all of these settings can be edited by passing
environment variables to the container when it is created.

When editing the config on bare-metal you can edit the config file directly. It can be found at
*dir_you_copied_repo_to*/simply_syslog/config/config.json

This is a quick reference guide to the configurations:

| Setting             | Explanation                                                                                               | Currently in use?                   | Allowed Values                                | Unit    |
|---------------------|-----------------------------------------------------------------------------------------------------------|-------------------------------------|-----------------------------------------------|---------|
| PROTOCOL            | The type of socket the server should listen on.                                                           | Partial, no support for TCP or BOTH | UDP                                           | N/A     |
| BIND_ADDRESS        | The address to have the server listen on. 0.0.0.0 will listen on all interfaces.                          | Yes                                 | Must be an IPv4 address                       | N/A     |
| UDP_PORT            | The UDP port that the server will listen on.                                                              | Yes                                 | Must be an integer value                      | N/A     |
| TCP_PORT            | The TCP port that the server will listen on.                                                              | No                                  | Must be an integer value                      | N/A     |
| MAX_TCP_CONNECTIONS | The max number of tcp connections that can be queued by the sever.                                        | No                                  | Must be an integer value                      | N/A     |
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

## Reporting Vulnerabilities:

If you find any vulnerabilities with this code please report the issue using the security section of the
GitHub repo (https://github.com/TheTurnnip/simply_syslog).

