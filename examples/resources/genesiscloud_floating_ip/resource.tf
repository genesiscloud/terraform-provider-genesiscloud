resource "genesiscloud_floating_ip" "floating_ip" {
  name        = "terraform-floating-ip"
  description = "The description for you terraform floating IP."
  region      = "NORD-NO-KRS-1"
  version     = "ipv4"
}
