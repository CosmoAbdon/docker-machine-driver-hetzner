# Hetzner Cloud Docker Machine Driver

[![Go Report Card](https://goreportcard.com/badge/github.com/CosmoAbdon/docker-machine-driver-hetzner)](https://goreportcard.com/report/github.com/CosmoAbdon/docker-machine-driver-hetzner)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go-CI](https://github.com/CosmoAbdon/docker-machine-driver-hetzner/actions/workflows/go.yml/badge.svg)](https://github.com/CosmoAbdon/docker-machine-driver-hetzner/actions/workflows/go.yml)

> This library adds support for creating [Docker machines](https://github.com/docker/machine) hosted on [Hetzner Cloud](https://www.hetzner.de/cloud), with full **Rancher/RKE2 compatibility**.

You need to create a project-specific access token under `Access` > `API Tokens` in the project control panel
and pass that to `docker-machine create` with the `--hetzner-api-token` option.

## Table of Contents

- [Rancher Users](#rancher-users)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Options](#options)
- [Building from source](#building-from-source)
- [Development](#development)
- [Changelog & Roadmap](#changelog--roadmap)
- [Credits](#credits)
- [License](#license)

## Rancher Users

**If you're using this driver with Rancher**, you'll need the UI extension for a proper integration experience:

👉 **[rancher-node-driver-hetzner](https://github.com/CosmoAbdon/rancher-node-driver-hetzner)** - Rancher UI Extension

The UI extension provides:
- Native Rancher interface for creating and managing Hetzner Cloud clusters
- Easy configuration of node pools, networks, firewalls, and more
- Seamless integration with Rancher's cluster management

**Quick Setup:**
1. Go to **Cluster Management** → **Drivers** → **Node Drivers**
2. Click **Add Node Driver**
3. Use the download URL from the [releases page](https://github.com/CosmoAbdon/docker-machine-driver-hetzner/releases)
4. Follow the [UI extension setup](https://github.com/CosmoAbdon/rancher-node-driver-hetzner) for the complete experience

## Features

- Full Docker Machine compatibility
- **Rancher/RKE2 integration** - works seamlessly as a node driver
- Flexible SSH key management (existing keys, generated keys, or both)
- Cloud-init user data support with YAML merging
- Private networking, firewalls, and volumes support
- Placement groups for high availability

## Installation

You can find sources and pre-compiled binaries [here](https://github.com/CosmoAbdon/docker-machine-driver-hetzner/releases).

```bash
# Download the binary (linux amd64 - check releases page for other platforms)
$ wget https://github.com/CosmoAbdon/docker-machine-driver-hetzner/releases/download/v1.0.0/docker-machine-driver-hetzner_1.0.0_linux_amd64.tar.gz
$ tar -xvf docker-machine-driver-hetzner_1.0.0_linux_amd64.tar.gz

# Make it executable and copy the binary in a directory accessible with your $PATH
$ chmod +x docker-machine-driver-hetzner
$ cp docker-machine-driver-hetzner /usr/local/bin/
```

## Usage

```bash
$ docker-machine create \
  --driver hetzner \
  --hetzner-api-token=your-api-token-here \
  some-machine
```

### Using environment variables

```bash
$ HETZNER_API_TOKEN=your-api-token-here \
  && HETZNER_IMAGE=ubuntu-24.04 \
  && docker-machine create \
     --driver hetzner \
     some-machine
```

### Using a custom storage driver

Modern images typically use overlay2 by default. If you need a specific storage driver:

```bash
$ docker-machine create \
  --engine-storage-driver overlay2 \
  --driver hetzner \
  --hetzner-api-token=your-api-token-here \
  some-machine
```

### Using Cloud-init

```bash
$ CLOUD_INIT_USER_DATA=`cat <<EOF
#cloud-config
write_files:
  - path: /test.txt
    content: |
      Here is a line.
      Another line is here.
EOF
`

$ docker-machine create \
  --driver hetzner \
  --hetzner-api-token=your-api-token-here \
  --hetzner-user-data="${CLOUD_INIT_USER_DATA}" \
  some-machine
```

### Merging Additional User Data

You can merge additional cloud-init configuration with base user data. This is particularly useful with Rancher, which injects its own user data:

```bash
$ docker-machine create \
  --driver hetzner \
  --hetzner-api-token=your-api-token-here \
  --hetzner-user-data-file=/path/to/base-config.yaml \
  --hetzner-additional-user-data="#cloud-config
packages:
  - vim
  - htop
runcmd:
  - echo 'Additional setup complete'" \
  some-machine
```

The additional user data is **merged** with the base configuration:
- Lists (like `packages`, `runcmd`) are combined, with additional data **prepended**
- Maps are merged recursively
- Scalar values from additional data override base values

### Using a snapshot

Assuming your snapshot ID is `424242`:

```bash
$ docker-machine create \
  --driver hetzner \
  --hetzner-api-token=your-api-token-here \
  --hetzner-image-id=424242 \
  some-machine
```

## Options

- `--hetzner-api-token`: **required**. Your project-specific access token for the Hetzner Cloud API.
- `--hetzner-image`: The name (or ID) of the Hetzner Cloud image to use, see [Images API](https://docs.hetzner.cloud/#images-get-all-images) for how to get a list (defaults to `ubuntu-24.04`). *Explicitly specifying an image is **strongly** recommended and will be **required from v3.0.0 onwards**.*
- `--hetzner-image-arch`: The architecture to use during image lookup, inferred from the server type if not explicitly given.
- `--hetzner-image-id`: The id of the Hetzner cloud image (or snapshot) to use, see [Images API](https://docs.hetzner.cloud/#images-get-all-images) for how to get a list (mutually excludes `--hetzner-image`).
- `--hetzner-server-type`: The type of the Hetzner Cloud server, see [Server Types API](https://docs.hetzner.cloud/#server-types-get-all-server-types) for how to get a list (defaults to `cpx22`).
- `--hetzner-server-location`: The location to create the server in, see [Locations API](https://docs.hetzner.cloud/#locations-get-all-locations) for how to get a list.
- `--hetzner-existing-key-path`: Use an existing (local) SSH key instead of generating a new keypair. If a remote key with a matching fingerprint exists, it will be used as if specified using `--hetzner-existing-key-id`, rather than uploading a new key.
- `--hetzner-existing-key-id`: Use an existing (remote) SSH key. Can be used **without** `--hetzner-existing-key-path` for Rancher/RKE2 compatibility - in this case, a local key will be generated and uploaded as an additional key to enable standalone SSH access.
- `--hetzner-additional-key`: Upload an additional public key associated with the server, or associate an existing one with the same fingerprint. Can be specified multiple times.
- `--hetzner-user-data`: Cloud-init based data, passed inline as-is.
- `--hetzner-user-data-file`: Cloud-init based data, read from passed file.
- `--hetzner-additional-user-data`: Additional cloud-init based data, passed inline. This content will be merged into the base user data YAML. Useful for injecting additional configuration. If duplicate keys exist, lists are combined (additional data prepended), maps are merged recursively, and scalars are overwritten.
- `--hetzner-volumes`: Volume IDs or names which should be attached to the server.
- `--hetzner-networks`: Network IDs or names which should be attached to the server private network interface.
- `--hetzner-use-private-network`: Use private network.
- `--hetzner-firewalls`: Firewall IDs or names which should be applied on the server.
- `--hetzner-server-label`: `key=value` pairs of additional metadata to assign to the server.
- `--hetzner-key-label`: `key=value` pairs of additional metadata to assign to SSH key (only applies if newly created).
- `--hetzner-placement-group`: Add to a placement group by name or ID; a spread-group will be created on demand if it does not exist.
- `--hetzner-auto-spread`: Add to a `docker-machine` provided `spread` group (mutually exclusive with `--hetzner-placement-group`).
- `--hetzner-ssh-user`: Change the default SSH-User.
- `--hetzner-ssh-port`: Change the default SSH-Port.
- `--hetzner-primary-ipv4/6`: Sets an existing primary IP (v4 or v6 respectively) for the server, as documented in [Networking](#networking).
- `--hetzner-wait-on-error`: Amount of seconds to wait on server creation failure (0/no wait by default).
- `--hetzner-wait-on-polling`: Amount of seconds to wait between requests when waiting for some state to change. (Default: 1 second)
- `--hetzner-wait-for-running-timeout`: Max amount of seconds to wait until a machine is running. (Default: 0/no timeout)

Please beware, that for options referring to entities by name, such as server locations and types, the names used by the API may differ from the ones
shown in the server creation UI. If server creation fails due to a failure to resolve such issues, try another variant of the name (e.g. lowercase,
kebab-case). As of writing, server types use lowercase (i.e. `cx21` instead of `CX21`) and locations use a three-letter abbreviation suffixed by 1
(i.e. `fsn1` instead of `Falkenstein`).

### Image selection

When `--hetzner-image-id` is passed, it will be used for lookup by ID as-is. No additional validation is performed, and it is mutually exclusive with
other `--hetzner-image*`-flags.

When `--hetzner-image` is passed, lookup will happen either by name or by ID as per Hetzner-supplied logic. The lookup mechanism will filter by image
architecture, which is usually inferred from the server type. One may explicitly specify it using `--hetzner-image-arch` in which case the user
supplied value will take precedence.

While there is currently a default image as fallback, this behaviour will be removed in a future version. Explicitly specifying an operating system
image is strongly recommended for new deployments, and will be mandatory in upcoming versions.

### Existing SSH keys

The driver supports flexible SSH key management for different use cases:

**Standalone usage with existing local key:**
```bash
$ docker-machine create \
  --driver hetzner \
  --hetzner-existing-key-path=~/.ssh/my-key \
  --hetzner-existing-key-id=12345 \
  some-machine
```
When you specify both `--hetzner-existing-key-path` and `--hetzner-existing-key-id`, the driver will:
1. Copy the local key pair to the machine's store
2. Verify the local key fingerprint matches the remote key
3. Use the existing remote key for server creation

**Rancher/RKE2 compatibility (existing key ID only):**
```bash
$ docker-machine create \
  --driver hetzner \
  --hetzner-existing-key-id=12345 \
  some-machine
```
When you specify only `--hetzner-existing-key-id` without a local key path:
1. The driver verifies the remote key exists
2. A new local key pair is generated
3. The generated key is uploaded as an **additional key** (`machine-name-local`)
4. The server is created with **both** keys: the existing key (for Rancher) and the generated key (for standalone SSH)

This enables both Rancher (which injects its key via cloud-init) and docker-machine to access the server via SSH.

**Standard usage (no existing key):**
```bash
$ docker-machine create \
  --driver hetzner \
  some-machine
```
The driver generates a new key pair, uploads it to Hetzner, and uses it for the server.

Note: The driver will attempt to delete linked keys during machine removal, unless `--hetzner-existing-key-id` was used during creation.

### Environment variables and default values

| CLI option                           | Environment variable               | Default                    |
| ------------------------------------ | ---------------------------------- | -------------------------- |
| **`--hetzner-api-token`**            | `HETZNER_API_TOKEN`                |                            |
| `--hetzner-image`                    | `HETZNER_IMAGE`                    | `ubuntu-24.04` as fallback |
| `--hetzner-image-arch`               | `HETZNER_IMAGE_ARCH`               | _(infer from server)_      |
| `--hetzner-image-id`                 | `HETZNER_IMAGE_ID`                 |                            |
| `--hetzner-server-type`              | `HETZNER_TYPE`                     | `cpx22`                    |
| `--hetzner-server-location`          | `HETZNER_LOCATION`                 | _(let Hetzner choose)_     |
| `--hetzner-existing-key-path`        | `HETZNER_EXISTING_KEY_PATH`        | _(generate new keypair)_   |
| `--hetzner-existing-key-id`          | `HETZNER_EXISTING_KEY_ID`          | 0 _(upload new key)_       |
| `--hetzner-additional-key`           | `HETZNER_ADDITIONAL_KEYS`          |                            |
| `--hetzner-user-data`                | `HETZNER_USER_DATA`                |                            |
| `--hetzner-user-data-file`           | `HETZNER_USER_DATA_FILE`           |                            |
| `--hetzner-additional-user-data`     | `HETZNER_ADDITIONAL_USER_DATA`     |                            |
| `--hetzner-networks`                 | `HETZNER_NETWORKS`                 |                            |
| `--hetzner-firewalls`                | `HETZNER_FIREWALLS`                |                            |
| `--hetzner-volumes`                  | `HETZNER_VOLUMES`                  |                            |
| `--hetzner-use-private-network`      | `HETZNER_USE_PRIVATE_NETWORK`      | false                      |
| `--hetzner-disable-public-ipv4`      | `HETZNER_DISABLE_PUBLIC_IPV4`      | false                      |
| `--hetzner-disable-public-ipv6`      | `HETZNER_DISABLE_PUBLIC_IPV6`      | false                      |
| `--hetzner-disable-public`           | `HETZNER_DISABLE_PUBLIC`           | false                      |
| `--hetzner-server-label`             | (inoperative)                      | `[]`                       |
| `--hetzner-key-label`                | (inoperative)                      | `[]`                       |
| `--hetzner-placement-group`          | `HETZNER_PLACEMENT_GROUP`          |                            |
| `--hetzner-auto-spread`              | `HETZNER_AUTO_SPREAD`              | false                      |
| `--hetzner-ssh-user`                 | `HETZNER_SSH_USER`                 | root                       |
| `--hetzner-ssh-port`                 | `HETZNER_SSH_PORT`                 | 22                         |
| `--hetzner-primary-ipv4`             | `HETZNER_PRIMARY_IPV4`             |                            |
| `--hetzner-primary-ipv6`             | `HETZNER_PRIMARY_IPV6`             |                            |
| `--hetzner-wait-on-error`            | `HETZNER_WAIT_ON_ERROR`            | 0                          |
| `--hetzner-wait-on-polling`          | `HETZNER_WAIT_ON_POLLING`          | 1                          |
| `--hetzner-wait-for-running-timeout` | `HETZNER_WAIT_FOR_RUNNING_TIMEOUT` | 0                          |

### Networking

Given `--hetzner-primary-ipv4` or `--hetzner-primary-ipv6`, the driver
attempts to set up machine creation with an existing [primary IP](https://docs.hetzner.com/cloud/servers/primary-ips/overview/)
as follows: If the passed argument parses to a valid IP address, the primary IP is resolved via address.
Otherwise, it is resolved in the default Hetzner Cloud API way (i.e. via ID and name as a fallback).

No address family validation is performed, so when specifying an IP address it is the user's responsibility to pass the
appropriate type. This also applies to any given preconditions regarding the state of the address being attached.

If no existing primary IPs are specified and public address creation is not disabled for a given address family, a new
primary IP will be auto-generated by default. Primary IPs created in that fashion will exhibit whatever default behavior
Hetzner assigns them at the given time, so users should take care what retention flags etc. are being set.

When disabling all public IPs, `--hetzner-use-private-network` must be given.
`--hetzner-disable-public` will take care of that, and behaves as if
`--hetzner-disable-public-ipv4 --hetzner-disable-public-ipv6 --hetzner-use-private-network`
were given.
Using `--hetzner-use-private-network` implicitly or explicitly requires at least one `--hetzner-network`
to be given.

## Building from source

Use an up-to-date version of [Go](https://golang.org/dl) (1.24+) to use Go Modules.

To use the driver, you can download the sources and build it locally:

```shell
# Clone the repository
$ git clone https://github.com/CosmoAbdon/docker-machine-driver-hetzner.git
$ cd docker-machine-driver-hetzner

# Build the binary
$ make build

# Or manually:
$ go build -o docker-machine-driver-hetzner .

# Make the binary accessible to docker-machine
$ cp docker-machine-driver-hetzner /usr/local/bin/
```

## Development

Fork this repository, yielding `github.com/<yourAccount>/docker-machine-driver-hetzner`.

```shell
# Clone your fork
$ git clone https://github.com/<yourAccount>/docker-machine-driver-hetzner.git
$ cd docker-machine-driver-hetzner

# Build and test
$ make build
$ make test

# Make docker-machine output help including hetzner-specific options
$ docker-machine create --driver hetzner
```

## Changelog & Roadmap

### 1.0.0 (Current)

This is the first release of the CosmoAbdon fork, featuring full Rancher/RKE2 compatibility.

**New Features:**
- **Rancher/RKE2 compatibility**: `--hetzner-existing-key-id` can now be used without `--hetzner-existing-key-path`
- **User data merging**: New `--hetzner-additional-user-data` flag for merging cloud-init configurations
- **UI Extension**: Companion [Rancher UI Extension](https://github.com/CosmoAbdon/rancher-node-driver-hetzner) for seamless cluster management

**Changes from upstream:**
- Default image updated to `ubuntu-24.04`
- Default server type updated to `cpx22`
- Internal code refactoring with improved package organization
- Empty network/firewall/volume values are now gracefully ignored (fixes Rancher UI compatibility)

### Planned: 1.x (Quick Wins)

Small improvements that enhance the driver without breaking changes:

- [ ] `--hetzner-automount` - Automatically mount attached volumes
- [ ] `--hetzner-start-after-create` - Control whether server starts immediately after creation
- [ ] Auto-create Primary IP if not specified (instead of relying on Hetzner defaults)
- [ ] Improved error messages with more context for better debugging

### Planned: 2.0.0

**Breaking Changes:**
- `--hetzner-user-data-from-file` will be removed entirely, including its fallback behavior
- `--hetzner-disable-public-4`/`--hetzner-disable-public-6` will be removed entirely, including their fallback behavior
- Not specifying `--hetzner-image` will generate a warning stating 'use of default image is DEPRECATED'

**New Features:**
- [ ] **Auto-create Network & Subnet**
  - `--hetzner-network-cidr` - Create a new network with specified CIDR if it doesn't exist
  - `--hetzner-subnet-cidr` - Create subnet within the network
  - `--hetzner-network-zone` - Specify network zone (eu-central, us-east, etc.)

- [ ] **Auto-create Firewall with basic rules**
  - `--hetzner-firewall-allow-ssh` - Create firewall rule allowing SSH (port 22)
  - `--hetzner-firewall-allow-icmp` - Create firewall rule allowing ICMP (ping)
  - `--hetzner-firewall-allow-ports` - Custom ports to allow (e.g., "80,443,6443")

- [ ] **Auto-create Volume**
  - `--hetzner-volume-size` - Create and attach a new volume with specified size (GB)
  - `--hetzner-volume-format` - Filesystem format (ext4, xfs)
  - `--hetzner-volume-automount` - Mount point for the created volume

- [ ] **Load Balancer Integration**
  - Attach servers to existing load balancers during creation

### Planned: 3.0.0

**Breaking Changes:**
- Specifying `--hetzner-image` will be mandatory, and a default image will no longer be provided

---

Contributions are welcome! Feel free to open issues or pull requests for any of these features.

## Credits

This project is a fork of [JonasProgrammer/docker-machine-driver-hetzner](https://github.com/JonasProgrammer/docker-machine-driver-hetzner), with additional contributions from:

- **[JonasProgrammer](https://github.com/JonasProgrammer)** - Original author and maintainer
- **[mxschmitt](https://github.com/mxschmitt)** - Core contributions and maintenance
- **[bluquist](https://github.com/bluquist)** - Rancher/RKE2 compatibility improvements
- **[CosmoAbdon](https://github.com/CosmoAbdon)** - Current maintainer, Rancher integration, code refactoring

## License

MIT License - see [LICENSE](LICENSE) for details.
