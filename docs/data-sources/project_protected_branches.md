# gitlab\_project\_protected\_branches

Provides details about all protected branches in a given project.

## Example Usage

```hcl
data "gitlab_project_protected_branches" "example" {
  project_id = 30
}
```

```hcl
data "gitlab_project_protected_branches" "example" {
  project_id = "foo/bar/baz"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) The integer or path with namespace that uniquely identifies the project.

## Attributes Reference

The following attributes are exported:

* `protected_branches` - A list of protected branches, as defined below.

## Nested Blocks

### `protected_branches`

* `id` - The ID of the protected branch.

* `name` - The name of the protected branch.

* `push_access_levels`, `merge_access_levels`, `unprotect_access_levels` - Each block contains a list of which access levels, users or groups are allowed to perform the respective actions (documented below).

* `code_owner_approval_required` - Reject code pushes that change files listed in the CODEOWNERS file.

### `push_access_levels`, `merge_access_levels`, `unprotect_access_levels`

#### Attributes

* `access_level` - The access level allowed to perform the respective action (shows as 40 - "maintainer" if `user_id` or `group_id` are present).

* `access_level_description` - A description of the allowed access level(s), or the name of the user or group if `user_id` or `group_id` are present.

* `user_id` - If present, indicates that the user is allowed to perform the respective action.

* `group_id` - If present, indicates that the group is allowed to perform the respective action.
