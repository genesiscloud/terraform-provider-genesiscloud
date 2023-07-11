package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccInstanceResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "genesiscloud_instance" "test" {
  name = %[1]q
}
`, name)
}

func TestAccInstanceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccInstanceResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("genesiscloud_instance.test", "id", "instance-id"),
					resource.TestCheckResourceAttr("genesiscloud_instance.test", "name", "one"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "genesiscloud_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccInstanceResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("genesiscloud_instance.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
