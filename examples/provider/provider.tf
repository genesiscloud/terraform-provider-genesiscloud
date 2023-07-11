terraform {
  required_providers {
    genesiscloud = {
      source = "genesiscloud/genesiscloud"
      # version = "..."
    }
  }
}

provider "genesiscloud" {
  # optional configuration...

  # set GENESISCLOUD_TOKEN env var or:
  # token = "..."
}

# Create an instance:

locals {
  region = "ARC-IS-HAF-1"
}

data "genesiscloud_images" "base-os" {
  filter = {
    type   = "base-os"
    region = local.region
  }
}

locals {
  image_id = data.genesiscloud_images.base-os.images[index(data.genesiscloud_images.base-os.images.*.name, "Ubuntu 20.04")].id
}

resource "genesiscloud_ssh_key" "alice" {
  name       = "alice"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBOpdKM8wSI07+PO4xLDL7zW/kNWGbdFXeHyBU1TRlBn alice@example.com"
}

resource "genesiscloud_security_group" "allow-ssh" {
  name   = "allow-ssh"
  region = local.region
  rules = [
    {
      direction      = "ingress"
      protocol       = "tcp"
      port_range_min = 22
      port_range_max = 22
    },
  ]
}

resource "genesiscloud_security_group" "allow-http" {
  name   = "allow-http"
  region = local.region
  rules = [
    {
      direction      = "ingress"
      protocol       = "tcp"
      port_range_min = 80
      port_range_max = 80
    }
  ]
}

resource "genesiscloud_security_group" "allow-https" {
  name   = "allow-https"
  region = local.region
  rules = [
    {
      direction      = "ingress"
      protocol       = "tcp"
      port_range_min = 443
      port_range_max = 443
    }
  ]
}

resource "genesiscloud_instance" "instance" {
  name   = "terraform-instance"
  region = local.region

  image_id = local.image_id
  type     = "vcpu-4_memory-12g_disk-80g_nvidia3080-1"

  ssh_key_ids = [
    genesiscloud_ssh_key.alice.id,
  ]

  security_group_ids = [
    genesiscloud_security_group.allow-ssh.id,
    genesiscloud_security_group.allow-http.id,
    genesiscloud_security_group.allow-https.id,
  ]

  metadata = {
    startup_script = <<EOF
#!/bin/bash
set -eo pipefail

# Add startup script

EOF

  }
}

output "connect" {
  value = "ssh ubuntu@${genesiscloud_instance.instance.public_ip}"
}
