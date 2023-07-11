resource "genesiscloud_ssh_key" "example" {
  name       = "example"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBOpdKM8wSI07+PO4xLDL7zW/kNWGbdFXeHyBU1TRlBn alice@example.com"
}
