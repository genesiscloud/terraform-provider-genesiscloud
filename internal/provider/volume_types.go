package provider

import (
	"context"
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VolumeResourceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Description The human-readable description for the volume.
	Description types.String `tfsdk:"description"`

	// Id The unique ID of the volume.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the volume.
	Name types.String `tfsdk:"name"`

	// Region The region identifier.
	Region types.String `tfsdk:"region"`

	// Size The storage size of this volume given in GiB.
	Size types.Int64 `tfsdk:"size"`

	// Status The volume status.
	Status types.String `tfsdk:"status"`

	// Type The storage type of the volume.
	Type types.String `tfsdk:"type"`

	// Internal

	// RetainOnDelete Flag to retain the volume when the resource is deleted. It has to be deleted manually.
	RetainOnDelete types.Bool `tfsdk:"retain_on_delete"`

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *VolumeResourceModel) PopulateFromClientResponse(ctx context.Context, volume *genesiscloud.Volume) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(volume.CreatedAt.Format(time.RFC3339))
	data.Description = types.StringValue(volume.Description)
	data.Id = types.StringValue(volume.Id)
	data.Name = types.StringValue(volume.Name)
	data.Region = types.StringValue(string(volume.Region))
	data.Size = types.Int64Value(int64(volume.Size))
	data.Status = types.StringValue(string(volume.Status))
	data.Type = types.StringValue(string(volume.Type))

	return
}
