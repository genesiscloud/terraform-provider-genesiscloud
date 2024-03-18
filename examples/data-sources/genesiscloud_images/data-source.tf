data "genesiscloud_images" "cloud-images" {
  filter = {
    type = "cloud-image"
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
