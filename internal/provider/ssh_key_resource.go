package provider

import (
	"context"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &SSHKeyResource{}
	_ resource.ResourceWithConfigure   = &SSHKeyResource{}
	_ resource.ResourceWithImportState = &SSHKeyResource{}
)

func NewSSHKeyResource() resource.Resource {
	return &SSHKeyResource{}
}

// SSHKeyResource defines the resource implementation.
type SSHKeyResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

type SSHKeyResourceModel struct {
	SSHKeyModel

	// Internal

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *SSHKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *SSHKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SSH key resource",

		Attributes: map[string]schema.Attribute{
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this SSH key was created in RFC 3339.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"fingerprint": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The fingerprint of the SSH key.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the SSH key.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the SSH key.",
				Required:            true,
			}),
			"public_key": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "SSH public key.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			}),

			// Internal
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *SSHKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SSHKeyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Create)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	body := genesiscloud.CreateSSHKeyJSONRequestBody{}

	body.Name = data.Name.ValueString()
	body.Value = data.PublicKey.ValueString()

	response, err := r.client.CreateSSHKeyWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create ssh_key", err))
		return
	}

	sshkeyResponse := response.JSON201
	if sshkeyResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create ssh_key", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, sshkeyResponse)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a ssh_key resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SSHKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Read)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	sshKeyId := data.Id.ValueString()

	response, err := r.client.GetSSHKeyWithResponse(ctx, sshKeyId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read ssh_key", err))
		return
	}

	sshkeyResponse := response.JSON200
	if sshkeyResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read ssh_key", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, sshkeyResponse)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a ssh_key resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SSHKeyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Update)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	body := genesiscloud.UpdateSSHKeyJSONRequestBody{}

	body.Name = data.Name.ValueString()

	sshKeyId := data.Id.ValueString()

	response, err := r.client.UpdateSSHKeyWithResponse(ctx, sshKeyId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update ssh_key", err))
		return
	}

	sshkeyResponse := response.JSON200
	if sshkeyResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update ssh_key", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, sshkeyResponse)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a ssh_key resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SSHKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Delete)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	sshKeyId := data.Id.ValueString()

	response, err := r.client.DeleteSSHKeyWithResponse(ctx, sshKeyId)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete ssh_key", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete ssh_key", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}
}

func (r *SSHKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
