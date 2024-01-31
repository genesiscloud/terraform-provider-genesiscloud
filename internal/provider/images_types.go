package provider

import (
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ImagesFilterDataSourceModel struct {
	// Region Filter by the region identifier.
	Region types.String `tfsdk:"region"`

	// Type Filter by the kind of image.
	Type types.String `tfsdk:"type"`
}

// ImagesDataSourceModel describes the data source data model.
type ImagesDataSourceModel struct {
	Filter ImagesFilterDataSourceModel `tfsdk:"filter"`
	Images []ImageModel                `tfsdk:"images"`
	Id     types.String                `tfsdk:"id"` // placeholder

	// Internal

	// Timeouts The data source timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

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

	Slug types.String `tfsdk:"slug"`

	Versions []types.String `tfsdk:"versions"`
}

func (data *ImageModel) PopulateFromClientResponse(image *genesiscloud.Image) {
	data.CreatedAt = types.StringValue(image.CreatedAt.Format(time.RFC3339))
	data.Id = types.StringValue(image.Id)
	data.Name = types.StringValue(image.Name)
	data.Regions = nil
	for _, region := range image.Regions {
		data.Regions = append(data.Regions, types.StringValue(string(region)))
	}

	data.Type = types.StringValue(string(image.Type))

	if image.Slug != nil {
		data.Slug = types.StringValue(*image.Slug)
	}

	data.Versions = nil
	if image.Versions != nil {
		for _, version := range *image.Versions {
			data.Versions = append(data.Versions, types.StringValue(string(version)))
		}
	}
}
