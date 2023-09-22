package provider

import (
	"context"
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SSHKeyResourceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Fingerprint The fingerprint of the SSH key.
	Fingerprint types.String `tfsdk:"fingerprint"`

	// Id The unique ID of the SSH key.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the SSH key.
	Name types.String `tfsdk:"name"`

	// PublicKey SSH public key.
	PublicKey types.String `tfsdk:"public_key"`

	// Internal

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *SSHKeyResourceModel) PopulateFromClientResponse(ctx context.Context, sshKey *genesiscloud.SSHKey) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(sshKey.CreatedAt.Format(time.RFC3339))
	data.Fingerprint = types.StringValue(sshKey.Fingerprint)
	data.Id = types.StringValue(sshKey.Id)
	data.Name = types.StringValue(sshKey.Name)
	data.PublicKey = types.StringValue(sshKey.Value)

	return
}
