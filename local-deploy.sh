#!/bin/bash

set -eo pipefail

version="0.0.0-local"
platform="darwin_arm64" # use `uname -sm` to get the current platform

mkdir -p ~/.terraform.d/plugins/registry.terraform.io/genesiscloud/genesiscloud/$version/$platform
cp terraform-provider-genesiscloud ~/.terraform.d/plugins/registry.terraform.io/genesiscloud/genesiscloud/$version/$platform/terraform-provider-genesiscloud
