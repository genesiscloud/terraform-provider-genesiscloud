package provider

import (
	"context"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/datasourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ datasource.DataSource              = &ImagesDataSource{}
	_ datasource.DataSourceWithConfigure = &ImagesDataSource{}
)

func NewImagesDataSource() datasource.DataSource {
	return &ImagesDataSource{}
}

// ImagesDataSource defines the data source implementation.
type ImagesDataSource struct {
	DataSourceWithClient
	DataSourceWithTimeout
}

func (d *ImagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *ImagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Images data source",

		Attributes: map[string]schema.Attribute{
			"filter": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"type": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
						MarkdownDescription: "Filter by the kind of image.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(sliceStringify(genesiscloud.AllImageTypes)...),
						},
					}),
					"region": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
						MarkdownDescription: "Filter by the region identifier.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(sliceStringify(genesiscloud.AllRegions)...),
						},
					}),
				},
			},
			"images": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"created_at": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
							MarkdownDescription: "The timestamp when this image was created in RFC 3339.",
							Computed:            true,
						}),
						"id": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
							MarkdownDescription: "A unique number that can be used to identify and reference a specific image.",
							Computed:            true,
						}),
						"name": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
							MarkdownDescription: "The display name that has been given to an image.",
							Computed:            true,
						}),
						"regions": datasourceenhancer.Attribute(ctx, schema.SetAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "The list of regions in which this image can be used in.",
							Computed:            true,
						}),
						"type": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
							MarkdownDescription: "Describes the kind of image.",
							Computed:            true,
						}),
						"slug": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
							MarkdownDescription: "The image slug.",
							Computed:            true,
						}),
						"versions": datasourceenhancer.Attribute(ctx, schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "The list of versions if this is a cloud-image otherwise empty.",
							Computed:            true,
						}),
					},
				},
			},
			"id": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The ID of the data source itself.",
				Computed:            true,
			}),

			// Internal
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *ImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ImagesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := d.ContextWithTimeout(ctx, data.Timeouts.Read)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	filterType := pointer(genesiscloud.ImageType(data.Filter.Type.ValueString()))

	var filterRegion *genesiscloud.Region

	if !data.Filter.Region.IsNull() && !data.Filter.Region.IsUnknown() {
		filterRegion = pointer(genesiscloud.Region(data.Filter.Region.ValueString()))
	}

	for page := 1; ; page++ {
		response, err := d.client.ListImagesWithResponse(ctx, &genesiscloud.ListImagesParams{
			Page:    pointer(page),
			PerPage: pointer(100),
			Type:    filterType,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("read images", err))
			return
		}

		imagesResponse := response.JSON200
		if imagesResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read images", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		for _, image := range imagesResponse.Images {
			if filterRegion != nil {
				var found bool
				for _, region := range image.Regions {
					if region == *filterRegion {
						found = true
						break
					}
				}

				if !found {
					continue
				}
			}

			model := ImageModel{}
			model.PopulateFromClientResponse(&image)

			data.Images = append(data.Images, model)
		}

		if len(imagesResponse.Images) < 100 {
			// pagination done
			break
		}
	}

	data.Id = types.StringValue("none")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
