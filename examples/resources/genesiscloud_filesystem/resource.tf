resource "genesiscloud_filesystem" "example" {
  name        = "example"
  description = "Example filesystem"
  region      = "NORD-NO-KRS-1"
  size        = 50
  type        = "vast"
}
