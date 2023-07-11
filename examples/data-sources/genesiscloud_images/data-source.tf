data "genesiscloud_images" "all-images" {}

data "genesiscloud_images" "base-os-images" {
  filter = {
    type = "base-os"
  }
}

data "genesiscloud_images" "snapshots" {
  filter = {
    type   = "snapshot"
    region = "ARC-IS-HAF-1"
  }
}

data "genesiscloud_images" "preconfigured-images" {
  filter = {
    type = "preconfigured"
  }
}
