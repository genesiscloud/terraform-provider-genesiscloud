resource "genesiscloud_floating_ip" "floating_ip" {
  name    = "terraform-floating-ip"
  region  = "NORD-NO-KRS-1"
  version = "ipv4"
}
