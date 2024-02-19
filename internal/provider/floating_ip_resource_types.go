package provider

import (
	"context"
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FloatingIPResourceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Id The unique ID of the Floating IP.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the Floating IP.
	Name types.String `tfsdk:"name"`

	// IpAddress The IP address of the floating IP.
	IpAddress types.String `tfsdk:"ip_address"`

	// IsPublic A boolean value indicating whether the floating IP is public or private.
	IsPublic types.Bool `tfsdk:"is_public"`

	// Region The region identifier.
	Region types.String `tfsdk:"region"`

	UpdatedAt types.String `tfsdk:"updated_at"`

	// Status The instance status
	Status types.String `tfsdk:"status"`

	// Description The human-readable description for the floating IP.
	Description types.String `tfsdk:"description"`

	// Version The IP version of the floating IP.
	Version types.String `tfsdk:"version"`

	// Internal

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *FloatingIPResourceModel) PopulateFromClientResponse(ctx context.Context, floatingIP *genesiscloud.FloatingIP) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(floatingIP.CreatedAt.Format(time.RFC3339))
	data.Id = types.StringValue(floatingIP.Id)
	data.Name = types.StringValue(floatingIP.Name)
	data.IsPublic = types.BoolValue(floatingIP.IsPublic)
	data.Region = types.StringValue(string(floatingIP.Region))
	data.UpdatedAt = types.StringValue(floatingIP.UpdatedAt.Format(time.RFC3339))
	data.Status = types.StringValue(string(floatingIP.Status))
	data.Version = types.StringValue(string(floatingIP.Version))
	data.Description = types.StringValue(floatingIP.Description)

	if floatingIP.IpAddress != nil {
		data.IpAddress = types.StringValue(*floatingIP.IpAddress)
	} else {
		data.IpAddress = types.StringNull()
	}
	return
}
