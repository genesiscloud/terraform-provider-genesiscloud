package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccFilesystemResourceConfig(name string, size int) string {
	return fmt.Sprintf(`
resource "genesiscloud_filesystem" "test" {
  name = %[1]q
  size = %[2]q
}
`, name, size)
}

func TestAccFilesystemResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccFilesystemResourceConfig("one", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("genesiscloud_filesystem.test", "id", "ssh-key-id"),
					resource.TestCheckResourceAttr("genesiscloud_filesystem.test", "name", "one"),
					resource.TestCheckResourceAttr("genesiscloud_filesystem.test", "size", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "genesiscloud_filesystem.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccFilesystemResourceConfig("two", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("genesiscloud_filesystem.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
