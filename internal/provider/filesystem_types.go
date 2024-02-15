package provider

import (
	"context"
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FilesystemResourceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Description The human-readable description for the filesystem.
	Description types.String `tfsdk:"description"`

	// Id The unique ID of the filesystem.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the filesystem.
	Name types.String `tfsdk:"name"`

	// MountEndpointRange The mount endpoint range for the filesystem.
	MountEndpointRange types.List `tfsdk:"mount_endpoint_range"`

	// MountBasePath The base path for the filesystem mount.
	MountBasePath types.String `tfsdk:"mount_base_path"`

	// Region The region identifier.
	Region types.String `tfsdk:"region"`

	// Size The storage size of this filesystem given in GiB.
	Size types.Int64 `tfsdk:"size"`

	// Status The filesystem status.
	Status types.String `tfsdk:"status"`

	// Type The storage type of the filesystem.
	Type types.String `tfsdk:"type"`

	// Internal

	// RetainOnDelete Flag to retain the filesystem when the resource is deleted. It has to be deleted manually.
	RetainOnDelete types.Bool `tfsdk:"retain_on_delete"`

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *FilesystemResourceModel) PopulateFromClientResponse(ctx context.Context, filesystem *genesiscloud.Filesystem) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(filesystem.CreatedAt.Format(time.RFC3339))
	data.Description = types.StringValue(filesystem.Description)
	data.Id = types.StringValue(filesystem.Id)
	data.Name = types.StringValue(filesystem.Name)

	if filesystem.MountBasePath != nil {
		data.MountBasePath = types.StringValue(*filesystem.MountBasePath)
	} else {
		data.MountBasePath = types.StringValue("")
	}

	data.MountEndpointRange, diag = types.ListValueFrom(ctx, types.StringType, filesystem.MountEndpointRange)
	data.Region = types.StringValue(string(filesystem.Region))
	data.Size = types.Int64Value(int64(filesystem.Size))
	data.Status = types.StringValue(string(filesystem.Status))
	data.Type = types.StringValue(string(filesystem.Type))

	return
}
