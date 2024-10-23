package provider

import (
	"context"
	"fmt"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/genesiscloud/terraform-provider-genesiscloud/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &InstanceStatusResource{}
	_ resource.ResourceWithConfigure   = &InstanceStatusResource{}
	_ resource.ResourceWithImportState = &InstanceStatusResource{}
)

func NewInstanceStatusResource() resource.Resource {
	return &InstanceStatusResource{}
}

// InstanceStatusResource defines the resource implementation.
type InstanceStatusResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

func (r *InstanceStatusResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_status"
}

func (r *InstanceStatusResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "InstanceStatus resource",

		Attributes: map[string]schema.Attribute{
			"instance_id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The id of the instance this refers to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			}),
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The target instance status.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(genesiscloud.InstanceStatusActive), string(genesiscloud.InstanceStatusStopped)),
				},
			}),

			// Internal
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *InstanceStatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceStatusResourceModel

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

	instanceId := data.InstanceId.ValueString()
	targetStatus := genesiscloud.InstanceStatus(data.Status.ValueString())

	var instanceAction genesiscloud.InstanceAction

	{
		response, err := r.client.GetInstanceWithResponse(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("get instance status", err))
			return
		}

		instanceResponse := response.JSON200
		if instanceResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("get instance status", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &instanceResponse.Instance)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Save data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if instanceResponse.Instance.Status == targetStatus {
			return
		}

		if targetStatus == genesiscloud.InstanceStatusActive &&
			instanceResponse.Instance.Status == genesiscloud.InstanceStatusStopped {

			instanceAction = genesiscloud.InstanceActionStart
		} else if targetStatus == genesiscloud.InstanceStatusStopped &&
			(instanceResponse.Instance.Status == genesiscloud.InstanceStatusActive ||
				instanceResponse.Instance.Status == genesiscloud.InstanceStatusStarting ||
				instanceResponse.Instance.Status == genesiscloud.InstanceStatusError) {

			instanceAction = genesiscloud.InstanceActionStop
		} else {
			resp.Diagnostics.AddError("Cannot transition instance status",
				fmt.Sprintf("The instance resource with id %q cannot be transitioned from %q status to %q status. If the current status is transient (ending in 'ing' e.g. stopping) waiting a bit is usually enough.",
					instanceId, instanceResponse.Instance.Status, targetStatus))
			return
		}
	}

	body := genesiscloud.PerformInstanceActionJSONRequestBody{}
	body.Action = instanceAction

	response, err := r.client.PerformInstanceActionWithResponse(ctx, instanceId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("perform instance action", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("perform instance action", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	tflog.Trace(ctx, "performed instance action", map[string]interface{}{"action": body.Action})

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling instance status", err))
			return
		}

		tflog.Trace(ctx, "polling instance status")

		response, err := r.client.GetInstanceWithResponse(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling instance status", err))
			return
		}

		instanceResponse := response.JSON200
		if instanceResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling instance status", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &instanceResponse.Instance)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Save data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if instanceResponse.Instance.Status == targetStatus {
			return
		}
	}
}

func (r *InstanceStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceStatusResourceModel

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

	instanceId := data.InstanceId.ValueString()

	response, err := r.client.GetInstanceWithResponse(ctx, instanceId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read instance status", err))
		return
	}

	instanceResponse := response.JSON200
	if instanceResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read instance status", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &instanceResponse.Instance)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read instance status resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceStatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InstanceStatusResourceModel

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

	instanceId := data.InstanceId.ValueString()
	targetStatus := genesiscloud.InstanceStatus(data.Status.ValueString())

	var instanceAction genesiscloud.InstanceAction

	{
		response, err := r.client.GetInstanceWithResponse(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("get instance status", err))
			return
		}

		instanceResponse := response.JSON200
		if instanceResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("get instance status", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &instanceResponse.Instance)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Save data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if instanceResponse.Instance.Status == targetStatus {
			return
		}

		if targetStatus == genesiscloud.InstanceStatusActive &&
			instanceResponse.Instance.Status == genesiscloud.InstanceStatusStopped {

			instanceAction = genesiscloud.InstanceActionStart
		} else if targetStatus == genesiscloud.InstanceStatusStopped &&
			(instanceResponse.Instance.Status == genesiscloud.InstanceStatusActive ||
				instanceResponse.Instance.Status == genesiscloud.InstanceStatusStarting ||
				instanceResponse.Instance.Status == genesiscloud.InstanceStatusError) {

			instanceAction = genesiscloud.InstanceActionStop
		} else {
			resp.Diagnostics.AddError("Cannot transition instance status",
				fmt.Sprintf("The instance resource with id %q cannot be transitioned from %q status to %q status. If the current status is transient (ending in 'ing' e.g. stopping) waiting a bit is usually enough.",
					instanceId, instanceResponse.Instance.Status, targetStatus))
			return
		}
	}

	body := genesiscloud.PerformInstanceActionJSONRequestBody{}
	body.Action = instanceAction

	response, err := r.client.PerformInstanceActionWithResponse(ctx, instanceId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("perform instance action", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("perform instance action", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	tflog.Trace(ctx, "performed instance action", map[string]interface{}{"action": body.Action})

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling instance status", err))
			return
		}

		tflog.Trace(ctx, "polling instance status")

		response, err := r.client.GetInstanceWithResponse(ctx, instanceId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling instance status", err))
			return
		}

		instanceResponse := response.JSON200
		if instanceResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling instance status", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &instanceResponse.Instance)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Save data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if instanceResponse.Instance.Status == targetStatus {
			return
		}
	}
}

func (r *InstanceStatusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceStatusResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// the resource is not real so its a noop
}

func (r *InstanceStatusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("instance_id"), req, resp)
}
