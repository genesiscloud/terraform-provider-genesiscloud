resource "genesiscloud_instance" "example" {
  name   = "example"
  region = "ARC-IS-HAF-1"

  image = "my-image-id"
  type  = "vcpu-2_memory-4g_disk-80g"

  ssh_key_ids = [
    "my-ssh-key-id"
  ]
}
