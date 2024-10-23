#!/bin/bash

set -eo pipefail

version="0.0.0-local"
platform="$(go env GOOS)_$(go env GOARCH)"

mkdir -p ~/.terraform.d/plugins/registry.terraform.io/genesiscloud/genesiscloud/$version/$platform
cp terraform-provider-genesiscloud ~/.terraform.d/plugins/registry.terraform.io/genesiscloud/genesiscloud/$version/$platform/terraform-provider-genesiscloud

# Cleanup:
# rm -fr ~/.terraform.d/plugins/registry.terraform.io/genesiscloud/
