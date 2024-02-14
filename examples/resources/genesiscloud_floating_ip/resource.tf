resource "genesiscloud_floating_ip" "floating_ip" {
  name        = "terraform-floating-ip"
  description = "The description for you terraform floating IP."
  region      = "ARC-IS-HAF-1"
  version     = "ipv4"
}
