package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccSSHKeyResourceConfig(name, publicKey string) string {
	return fmt.Sprintf(`
resource "genesiscloud_ssh_key" "test" {
  name       = %[1]q
  public_key = %[2]q
}
`, name, publicKey)
}

const samplePublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBOpdKM8wSI07+PO4xLDL7zW/kNWGbdFXeHyBU1TRlBn alice@example.com"

func TestAccSSHKeyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccSSHKeyResourceConfig("one", samplePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("genesiscloud_ssh_key.test", "id", "ssh-key-id"),
					resource.TestCheckResourceAttr("genesiscloud_ssh_key.test", "name", "one"),
					resource.TestCheckResourceAttr("genesiscloud_ssh_key.test", "public_key", samplePublicKey),
				),
			},
			// ImportState testing
			{
				ResourceName:      "genesiscloud_ssh_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccSSHKeyResourceConfig("two", samplePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("genesiscloud_ssh_key.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
