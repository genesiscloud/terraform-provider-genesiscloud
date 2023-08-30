package provider

import (
	"context"
	"fmt"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/defaultplanmodifier"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &SnapshotResource{}
	_ resource.ResourceWithConfigure   = &SnapshotResource{}
	_ resource.ResourceWithImportState = &SnapshotResource{}
)

func NewSnapshotResource() resource.Resource {
	return &SnapshotResource{}
}

// SnapshotResource defines the resource implementation.
type SnapshotResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

type SnapshotResourceModel struct {
	SnapshotModel

	// Internal

	// RetainOnDelete Flag to retain the snapshot when the resource is deleted. It has to be deleted manually.
	RetainOnDelete types.Bool `tfsdk:"retain_on_delete"`

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *SnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot"
}

func (r *SnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Snapshot resource",

		Attributes: map[string]schema.Attribute{
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this snapshot was created in RFC 3339.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the snapshot.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the snapshot.",
				Required:            true,
			}),
			"region": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The region identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"instance_id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The id of the instance to snapshot.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			}),
			"size": resourceenhancer.Attribute(ctx, schema.Int64Attribute{
				MarkdownDescription: "The storage size of this snapshot given in bytes.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(), // immutable
				},
			}),
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The snapshot status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // immutable
				},
			}),

			// Internal
			"retain_on_delete": resourceenhancer.Attribute(ctx, schema.BoolAttribute{
				MarkdownDescription: "Flag to retain the snapshot when the resource is deleted.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					defaultplanmodifier.Bool(false),
				},
			}),

			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *SnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SnapshotResourceModel

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

	body := genesiscloud.CreateSnapshotJSONRequestBody{}
	body.Name = data.Name.ValueString()

	instanceId := data.InstanceId.ValueString()

	response, err := r.client.CreateSnapshotWithResponse(ctx, instanceId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create snapshot", err))
		return
	}

	snapshotResponse := response.JSON201
	if snapshotResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create snapshot", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &snapshotResponse.Snapshot)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a snapshot resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshotId := snapshotResponse.Snapshot.Id

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling snapshot", err))
			return
		}

		tflog.Trace(ctx, "polling a snapshot resource")

		response, err := r.client.GetSnapshotWithResponse(ctx, snapshotId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling snapshot", err))
			return
		}

		snapshotResponse := response.JSON200
		if snapshotResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling snapshot", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		status := snapshotResponse.Snapshot.Status
		if status == "created" || status == "active" || status == "error" {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &snapshotResponse.Snapshot)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == "error" {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling snapshot", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *SnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnapshotResourceModel

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

	snapshotId := data.Id.ValueString()

	response, err := r.client.GetSnapshotWithResponse(ctx, snapshotId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read snapshot", err))
		return
	}

	snapshotResponse := response.JSON200
	if snapshotResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read snapshot", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &snapshotResponse.Snapshot)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a snapshot resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SnapshotResourceModel

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

	body := genesiscloud.UpdateSnapshotJSONRequestBody{}

	body.Name = data.Name.ValueString()

	snapshotId := data.Id.ValueString()

	response, err := r.client.UpdateSnapshotWithResponse(ctx, snapshotId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update snapshot", err))
		return
	}

	snapshotResponse := response.JSON200
	if snapshotResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update snapshot", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &snapshotResponse.Snapshot)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a snapshot resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SnapshotResourceModel

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

	snapshotId := data.Id.ValueString()

	if data.RetainOnDelete.ValueBool() {
		resp.Diagnostics.AddWarning(
			"Snapshot is retained",
			fmt.Sprintf("The snapshot resource with id %q was deleted from the state but the snapshot is retained.", snapshotId),
		)
		return
	}

	response, err := r.client.DeleteSnapshotWithResponse(ctx, snapshotId)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete snapshot", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete snapshot", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling snapshot", err))
			return
		}

		tflog.Trace(ctx, "polling a snapshot resource")

		response, err := r.client.GetSnapshotWithResponse(ctx, snapshotId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling snapshot", err))
			return
		}

		if response.StatusCode() == 404 {
			return
		}
	}
}

func (r *SnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
