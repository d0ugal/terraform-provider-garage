# Terraform Provider for Garage

A lightweight Terraform provider for Garage storage using the v1 admin API.

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
      source  = "hoose/garage"
      version = "0.1.0"
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

This installs the provider to `~/.terraform.d/plugins/registry.terraform.io/hoose/garage/0.1.0/linux_amd64/`

## Development

```bash
# Build
make build

# Test
make test

# Clean
make clean
```


