package provider

import (
	"context"
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InstanceMetadataModel struct {
	// StartupScript A plain text bash script or "cloud-config" file that will be executed after the first instance boot.
	// It is limited to 64 KiB in size. You can use it to configure your instance, e.g. installing the **NVIDIA GPU driver**.
	// Learn more about [startup scripts and installing the GPU driver](https://support.genesiscloud.com/support/solutions/articles/47001122478).
	StartupScript types.String `tfsdk:"startup_script"`
}

type InstanceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Hostname The hostname of your instance.
	Hostname types.String `tfsdk:"hostname"`

	// Id The unique ID of the instance.
	Id types.String `tfsdk:"id"`

	// ImageId The image of the instance.
	ImageId types.String `tfsdk:"image_id"`

	// Metadata Option to provide metadata. Currently supported is `startup_script`.
	Metadata *InstanceMetadataModel `tfsdk:"metadata"`

	// Name The human-readable name for the instance.
	Name types.String `tfsdk:"name"`

	// Password The password to access the instance.
	// Your password must have a minimum length of 16 characters.
	// **Please Note**: Only one of `ssh_keys` or `password` can be provided.
	// Password is less secure - we recommend you use an SSH key-pair.
	Password types.String `tfsdk:"password"`

	// PlacementOption The placement option identifier in which instances are physically located relative to each other within a zone.
	PlacementOption types.String `tfsdk:"placement_option"`

	// PrivateIp The private IPv4 IP-Address (IPv4 address).
	PrivateIp types.String `tfsdk:"private_ip"`

	// PublicIp The public IPv4 IP-Address (IPv4 address).
	PublicIp types.String `tfsdk:"public_ip"`

	// PublicIpType When set to `static`, the instance's public IP will not change between start and stop actions.
	PublicIpType types.String `tfsdk:"public_ip_type"`

	// Region The region identifier.
	Region types.String `tfsdk:"region"`

	// SecurityGroupIds The security groups of the instance.
	SecurityGroupIds types.Set `tfsdk:"security_group_ids"`

	// SshKeyIds The ssh keys of the instance.
	SshKeyIds types.Set `tfsdk:"ssh_key_ids"`

	// Status The instance status
	Status types.String `tfsdk:"status"`

	// Type The instance type identifier.
	Type types.String `tfsdk:"type"`

	UpdatedAt types.String `tfsdk:"updated_at"`

	// VolumeIds The volumes of the instance
	VolumeIds types.Set `tfsdk:"volume_ids"`
}

func (data *InstanceModel) PopulateFromClientResponse(ctx context.Context, instance *genesiscloud.Instance) (diag diag.Diagnostics) {
	data.Id = types.StringValue(instance.Id)
	data.Name = types.StringValue(instance.Name)
	data.Hostname = types.StringValue(instance.Hostname)
	data.Type = types.StringValue(string(instance.Type))
	data.ImageId = types.StringValue(instance.Image.Id)

	volumeIds := make([]string, 0) // volumes do NOT support NULL
	for _, volume := range instance.Volumes {
		volumeIds = append(volumeIds, volume.Id)
	}
	data.VolumeIds, diag = types.SetValueFrom(ctx, types.StringType, volumeIds)
	if diag.HasError() {
		return
	}

	securityGroupIds := make([]string, 0) // security groups do NOT support NULL
	for _, securityGroup := range instance.SecurityGroups {
		securityGroupIds = append(securityGroupIds, securityGroup.Id)
	}
	data.SecurityGroupIds, diag = types.SetValueFrom(ctx, types.StringType, securityGroupIds)
	if diag.HasError() {
		return
	}

	var sshKeyIds []string // ssh-keys do support NULL
	for _, sshKey := range instance.SshKeys {
		sshKeyIds = append(sshKeyIds, sshKey.Id)
	}
	data.SshKeyIds, diag = types.SetValueFrom(ctx, types.StringType, sshKeyIds)
	if diag.HasError() {
		return
	}

	data.PlacementOption = types.StringValue(string(instance.PlacementOption))
	data.PrivateIp = types.StringValue(*instance.PrivateIp)
	data.PublicIp = types.StringValue(*instance.PublicIp)
	data.PublicIpType = types.StringValue(string(instance.PublicIpType))
	data.Region = types.StringValue(string(instance.Region))
	data.Status = types.StringValue(string(instance.Status))
	data.CreatedAt = types.StringValue(instance.CreatedAt.Format(time.RFC3339))
	data.UpdatedAt = types.StringValue(instance.UpdatedAt.Format(time.RFC3339))

	return
}
