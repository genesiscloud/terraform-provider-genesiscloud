package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccImagesDataSourceConfig = `
data "genesiscloud_images" "test" {
	type = "cloud-image"
}
`

func TestAccImagesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + testAccImagesDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of images returned
					resource.TestCheckResourceAttr("data.genesiscloud_images.test", "images.#", "3"),
					// Verify the first coffee to ensure all attributes are set
					resource.TestCheckResourceAttr("data.genesiscloud_images.test", "images.0.id", "todo-uuid"),
					resource.TestCheckResourceAttr("data.genesiscloud_images.test", "images.0.name", "todo-name"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.genesiscloud_images.test", "id", "none"),
				),
			},
		},
	})
}
