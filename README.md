# Terraform Provider for Garage

A lightweight Terraform provider for Garage storage using the v1 admin API.

[![CI](https://github.com/d0ugal/terraform-provider-garage/actions/workflows/ci.yml/badge.svg)](https://github.com/d0ugal/terraform-provider-garage/actions/workflows/ci.yml)

## Features

- **garage_key**: Create and manage access keys
- **garage_bucket**: Create and manage buckets
- **garage_bucket_key**: Manage key permissions on buckets

## Building

```bash
make build
make install
```

## Usage

```hcl
terraform {
  required_providers {
    garage = {
      source  = "d0ugal/garage"
      version = "0.0.1"
    }
  }
}

provider "garage" {
  scheme = "http"
  host   = "127.0.0.1:3903"
  token  = "your-admin-token"
}

resource "garage_key" "loki_key" {
  name = "loki-access-key"
}

resource "garage_bucket" "loki" {
  global_alias = "loki"
}

resource "garage_bucket_key" "loki_access" {
  bucket_id     = garage_bucket.loki.id
  access_key_id = garage_key.loki_key.access_key_id
  read          = true
  write         = true
  owner         = false
}
```

## Installation

After building, install to your local Terraform plugins directory:

```bash
make install
```

This installs the provider to `~/.terraform.d/plugins/registry.terraform.io/d0ugal/garage/0.0.1/linux_amd64/`

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (to build the provider plugin)
- Access to a Garage instance with admin API enabled

## Development

```bash
# Build
make build

# Test
make test

# Lint
make lint

# Clean
make clean
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


