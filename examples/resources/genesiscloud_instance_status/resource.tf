resource "genesiscloud_instance" "example" {
  name   = "example"
  region = "NORD-NO-KRS-1"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  ssh_key_ids = [
    "my-ssh-key-id"
  ]
}


resource "genesiscloud_instance_status" "example" {
  instance_id = genesiscloud_instance.example.id
  status      = "active"
}
