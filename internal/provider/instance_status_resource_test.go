package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccInstanceStatusResourceConfig(name string, size int) string {
	return fmt.Sprintf(`
resource "genesiscloud_instance_status" "test" {
  name = %[1]q
  size = %[2]q
}
`, name, size)
}

func TestAccInstanceStatusResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccInstanceStatusResourceConfig("one", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("genesiscloud_instance_status.test", "id", "ssh-key-id"),
					resource.TestCheckResourceAttr("genesiscloud_instance_status.test", "name", "one"),
					resource.TestCheckResourceAttr("genesiscloud_instance_status.test", "size", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "genesiscloud_instance_status.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccInstanceStatusResourceConfig("two", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("genesiscloud_instance_status.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
