<!--[metadata]>
+++
title = "Selectel Docker Machine Driver"
description = "selectel driver for docker machine"
keywords = ["machine, selectel, driver, docker"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->
# Selectel VPC driver for Docker Machine

[![Build Status](https://travis-ci.org/selectel/docker-machine-driver.svg?branch=master)](https://travis-ci.org/selectel/docker-machine-driver)

This is a driver plugin for Docker Machine. It allows you to
orchestrate virtual machines in Selectel Virtual Private Cloud (VPC)

## Installation

### Requirements

* Docker CLI (the daemon is not required) must be installed on the client machine. Tested with `version 18.03.1-ce, build 9ee9f40`
* [Docker Machine](https://docs.docker.com/machine/). Tested with `version: 0.14.0, build 89b8332`
* Access to an [Selectel](https://my.selectel.ru/vpc/access) Cloud.


### From Source

Clone this repo and run cmd which will install the driver into `/usr/local/bin`

```
$ make install
```

## Obtaining credentials
To use the driver you will need to complete those steps:
1. Create a project that will contain VMs[here](https://my.selectel.ru/vpc/projects)
2. Set CPU, RAM and volume quotas
3. Add floating IP to the project
4. Create a new user and set the project role [here](https://my.selectel.ru/vpc/users)

Then download the `rc.sh` [file](https://my.selectel.ru/vpc/access)
that contains env variables for managing OpenStack from CLI

🔥🔥 Attention! `rc.sh` does not contain information about `OS_AVAILABILITY_ZONE`

You should add it manually to the file or provide via `--os-availability-zone` as a СLI option.

### Example of `rc.sh`
```
#!/bin/bash

export OS_AUTH_URL="https://api.selvpc.ru/identity/v3"
export OS_IDENTITY_API_VERSION="3"
export OS_VOLUME_API_VERSION="2"

export OS_PROJECT_DOMAIN_NAME='your_domain_name_here'
export OS_PROJECT_ID='your_project_id_here'

export OS_REGION_NAME='ru-1'
export OS_AVAILABILITY_ZONE='ru-1b'

export OS_USER_DOMAIN_NAME='your_domain_name_here'
export OS_USERNAME='you_user_name_here'
export OS_PASSWORD='you_user_password_here'
```

## Usage
You may want to refer to the Docker Machine [official documentation](https://docs.docker.com/machine/) before using the driver.

Verify that the Selectel driver shows up in Docker Machine:

```bash
docker-machine create -d selectel --help
```
To create a Docker host, make sure that you have followed the instructions from `Obtaining credentials` section.

Create a VM with Docker and bind the local docker-client to the backend:
```bash
docker-machine create -d selectel you-server-name

eval $(docker-machine env you-server-name)
```

[![asciicast](https://asciinema.org/a/182929.png)](https://asciinema.org/a/182929)


### CLI Options, Environment variables and default values:

| CLI option                   | Default                     | Environment variable        | Description                                             |
|------------------------------|-----------------------------|-----------------------------|---------------------------------------------------------|
| `--os-auth-url`              |                             | `$OS_AUTH_URL`              | OpenStack authentication URL                            |
| `--os-domain-name`           |                             | `$OS_PROJECT_DOMAIN_NAME`   | OpenStack domain name (identity v3 only)                |
| `--os-flavor-id`             |                             | `$OS_FLAVOR_ID`             | OpenStack flavor id to use for the instance             |
| `--os-flavor-name`           |                             | `$OS_FLAVOR_NAME`           | OpenStack flavor name to use for the instance           |
| `--os-image-id`              |                             | `$OS_IMAGE_ID`              | OpenStack image id to use for the instance              |
| `--os-image-name`            | "Ubuntu 16.04 LTS 64-bit"   | `$OS_IMAGE_NAME`            | OpenStack flavor name to use for the instance           |
| `--os-net-id`                |                             | `$OS_NETWORK_ID`            | OpenStack network id the machine will be connected on   |
| `--os-project-id`            |                             | `$OS_PROJECT_ID`            | OpenStack project id                                    |
| `--os-region`                |                             | `$OS_REGION_NAME`           | OpenStack region name                                   |
| `--os-availability-zone`     |                             | `$OS_AVAILABILITY_ZONE`     | OpenStack availability zone                             |
| `--os-username`              |                             | `$OS_USERNAME`              | OpenStack username                                      |
| `--os-password`              |                             | `$OS_PASSWORD`              | OpenStack user password                                 |
| `--sel-cpu`                  | "1"                         | `$SEL_CPU_VALUE`            | Count of vCPU for server                                |
| `--sel-proxy`                |                             | `$SEL_PROXY`                | Proxy for the OS services                               |
| `--sel-ram`                  | "512"                       | `$SEL_RAM_VALUE`            | Count of RAM for server                                 |
| `--sel-server-name`          |                             | `$SEL_SERVER_NAME`          | Name of future server                                   |
| `--sel-ssh-pair-name`        | "docker-machine-key"        | `$SEL_SSH_PAIR_NAME`        | Existing keypair name                                   |
| `--sel-ssh-port`             | "22"                        | `$SEL_SSH_PORT`             | SSH port for connecting to the server                   |
| `--sel-ssh-private-key-path` |                             | `$SEL_SSH_PRIVATE_KEY_PATH` | Private keyfile to use for SSH (absolute path)          |
| `--sel-ssh-user`             | "root"                      | `$SEL_SSH_USER`             | SSH user for connecting to the server                   |
| `--sel-volume-name`          |                             | `$SEL_VOLUME_NAME`          | Name of the server volume                               |
| `--sel-volume-size`          | "5"                         | `$SEL_VOLUME_SIZE`          | Volume size                                             |
| `--sel-volume-type`          |                             | `$SEL_VOLUME_TYPE`          | Base volume type for server                             |

## Related links

- **Docker Machine**: https://docs.docker.com/machine/
- **Report bugs**: https://github.com/selectel/docker-machine-driver/issues
- **Contribute**: https://github.com/selectel/docker-machine-driver

## Authors

* Mikhail Kalinin ([@objque](https://github.com/objque))

## Contributing

We hope you'll get involved!

## License

Apache 2.0
