package provider

import (
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ImageModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Id A unique number that can be used to identify and reference a specific image.
	Id types.String `tfsdk:"id"`

	// Name The display name that has been given to an image.
	Name types.String `tfsdk:"name"`

	// Regions The set of regions in which this image can be used in.
	Regions []types.String `tfsdk:"regions"`

	// Type Describes the kind of image.
	Type types.String `tfsdk:"type"`
}

func (data *ImageModel) PopulateFromClientResponse(image *genesiscloud.ComputeV1Image) {
	data.CreatedAt = types.StringValue(image.CreatedAt.Format(time.RFC3339))
	data.Id = types.StringValue(image.Id)
	data.Name = types.StringValue(image.Name)
	data.Regions = nil
	for _, region := range image.Regions {
		data.Regions = append(data.Regions, types.StringValue(string(region)))
	}
	data.Type = types.StringValue(string(image.Type))
}
