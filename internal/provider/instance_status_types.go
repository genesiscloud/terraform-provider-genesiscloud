package provider

import (
	"context"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InstanceStatusResourceModel struct {
	// InstanceId The id of the instance this refers to.
	InstanceId types.String `tfsdk:"instance_id"`

	// Status The target instance status.
	Status types.String `tfsdk:"status"`

	// Internal

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *InstanceStatusResourceModel) PopulateFromClientResponse(ctx context.Context, instance *genesiscloud.Instance) (diag diag.Diagnostics) {
	data.InstanceId = types.StringValue(instance.Id)
	data.Status = types.StringValue(string(instance.Status))

	return
}
