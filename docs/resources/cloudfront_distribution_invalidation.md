---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "awsex_cloudfront_distribution_invalidation Resource - awsex"
subcategory: ""
description: |-
  
---

# awsex_cloudfront_distribution_invalidation (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `distribution_id` (String) The Cloudfront Distribution ID where an invalidation should be created.
- `paths` (Set of String) A list of paths to invalidate. Each path *must* start with `/`.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `triggers` (Map of String) A map of triggers that, when changed, will force Terraform to create a new invalidation.

### Read-Only

- `id` (String) The ID of the invalidation.
- `status` (String) The status of the invalidation.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
