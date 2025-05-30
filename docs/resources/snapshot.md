---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "genesiscloud_snapshot Resource - terraform-provider-genesiscloud"
subcategory: ""
description: |-
  Snapshot resource
---

# genesiscloud_snapshot (Resource)

Snapshot resource

## Example Usage

```terraform
resource "genesiscloud_instance" "target" {
  # ...
}

resource "genesiscloud_snapshot" "example" {
  name        = "example"
  instance_id = genesiscloud_instance.target.id

  retain_on_delete = true # optional
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The human-readable name for the snapshot.

### Optional

- `region` (String) The region identifier. Should only be explicity specified when using the 'source_snapshot_id'.
- `replicated_region` (String) Target region for snapshot replication. When specified, also creates a copy of the snapshot in the given region. If omitted, the snapshot exists only in the current region.
- `retain_on_delete` (Boolean) Flag to retain the snapshot when the resource is deleted.
  - Sets the default value "false" if the attribute is not set.
- `source_instance_id` (String) The id of the source instance from which this snapshot was derived.
  - If the value of this attribute changes, the resource will be replaced.
- `source_snapshot_id` (String) The id of the source snapshot from which this snapsot was derived.
  - If the value of this attribute changes, the resource will be replaced.
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `created_at` (String) The timestamp when this snapshot was created in RFC 3339.
- `id` (String) The unique ID of the snapshot.
- `size` (Number) The storage size of this snapshot given in GiB.
- `status` (String) The snapshot status.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

Import is supported using the following syntax:

```shell
terraform import genesiscloud_snapshot.example 18efeec8-94f0-4776-8ff2-5e9b49c74608
```
