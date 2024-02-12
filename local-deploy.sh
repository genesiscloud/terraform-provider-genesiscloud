#!/bin/bash

set -eo pipefail

version="0.0.0-local"
platform="darwin_arm64" # check for your OS using `uname -sm`

mkdir -p ~/.terraform.d/plugins/registry.terraform.io/genesiscloud/genesiscloud/$version/$platform
cp terraform-provider-genesiscloud ~/.terraform.d/plugins/registry.terraform.io/genesiscloud/genesiscloud/$version/$platform/terraform-provider-genesiscloud
