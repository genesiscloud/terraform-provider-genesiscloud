package provider

import (
	"context"
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SnapshotResourceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Id The unique ID of the snapshot.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the snapshot.
	Name types.String `tfsdk:"name"`

	// Region The region identifier.
	Region types.String `tfsdk:"region"`

	// SourceInstanceId The id of the source instance from which this snapshot was derived.
	SourceInstanceId types.String `tfsdk:"source_instance_id"`

	// SourceSnapshotId The id of the source snapshot from which this snapsot was derived.
	SourceSnapshotId types.String `tfsdk:"source_snapshot_id"`

	// ReplicatedRegion The region identifier when the snapshot should be replicated.
	ReplicatedRegion types.String `tfsdk:"replicated_region"`

	// Size The storage size of this snapshot given in GiB.
	Size types.Int64 `tfsdk:"size"`

	// Status The snapshot status.
	Status types.String `tfsdk:"status"`

	// Internal

	// RetainOnDelete Flag to retain the snapshot when the resource is deleted. It has to be deleted manually.
	RetainOnDelete types.Bool `tfsdk:"retain_on_delete"`

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *SnapshotResourceModel) PopulateFromClientResponse(ctx context.Context, snapshot *genesiscloud.Snapshot) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(snapshot.CreatedAt.Format(time.RFC3339))
	data.Id = types.StringValue(snapshot.Id)
	data.Name = types.StringValue(snapshot.Name)
	data.Region = types.StringValue(string(snapshot.Region))
	data.Size = types.Int64Value(int64(snapshot.Size))
	data.Status = types.StringValue(string(snapshot.Status))

	if snapshot.SourceInstanceId != nil {
		data.SourceInstanceId = types.StringValue(*snapshot.SourceInstanceId)
	}
	if snapshot.SourceSnapshotId != nil {
		data.SourceSnapshotId = types.StringValue(*snapshot.SourceSnapshotId)
	}

	return
}
