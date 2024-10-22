data "genesiscloud_images" "cloud-images" {
  filter = {
    type = "cloud-image"
  }
}

data "genesiscloud_images" "snapshots" {
  filter = {
    type   = "snapshot"
    region = "NORD-NO-KRS-1"
  }
}

data "genesiscloud_images" "preconfigured-images" {
  filter = {
    type = "preconfigured"
  }
}
