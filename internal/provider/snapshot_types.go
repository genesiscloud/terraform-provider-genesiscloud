package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SnapshotModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Id The unique ID of the snapshot.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the snapshot.
	Name types.String `tfsdk:"name"`

	// Region The region identifier.
	Region types.String `tfsdk:"region"`

	// InstanceId The id of the instance that was snapshotted.
	InstanceId types.String `tfsdk:"instance_id"`

	// Size The storage size of this snapshot given in bytes.
	Size types.Int64 `tfsdk:"size"`

	// Status The snapshot status.
	Status types.String `tfsdk:"status"`
}

func (data *SnapshotModel) PopulateFromClientResponse(ctx context.Context, snapshot *genesiscloud.Snapshot) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(snapshot.CreatedAt.Format(time.RFC3339))
	data.Id = types.StringValue(snapshot.Id)
	data.Name = types.StringValue(snapshot.Name)
	data.Region = types.StringValue(string(snapshot.Region))
	data.InstanceId = types.StringValue(string(snapshot.ResourceId))

	data.Size = types.Int64Null()
	if snapshot.Size != nil {
		i, err := strconv.ParseInt(*snapshot.Size, 10, 64)
		if err != nil {
			diag.AddAttributeError(
				path.Root("size"),
				"Unmarshalling failed",
				fmt.Sprintf("Failed to unmarshal BigInt response: %q", *snapshot.Size),
			)
			return
		}

		data.Size = types.Int64Value(i)
	}

	data.Status = types.StringValue(string(snapshot.Status))

	return
}
