resource "genesiscloud_instance" "target" {
  # ...
}

resource "genesiscloud_snapshot" "example" {
  name        = "example"
  instance_id = genesiscloud_instance.target.id

  retain_on_delete = true # optional
}
