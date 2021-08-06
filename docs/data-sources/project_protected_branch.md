# gitlab\_project\_protected\_branch

Provides details about a specific protected branch in a given project.

## Example Usage

```hcl
data "gitlab_project_protected_branch" "example" {
  project_id = 30
  name       = "main"
}
```

```hcl
data "gitlab_project_protected_branch" "example" {
  project_id = "foo/bar/baz"
  name       = "main"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) The integer or path with namespace that uniquely identifies the project.

* `name` - (Required) The name of the protected branch.

## Attributes Reference

The following attributes are exported:

* `push_access_levels`, `merge_access_levels`, `unprotect_access_levels` - Each block contains a list of which access levels, users or groups are allowed to perform the respective actions (documented below).

* `code_owner_approval_required` - Reject code pushes that change files listed in the CODEOWNERS file.

## Nested Blocks

### `push_access_levels`, `merge_access_levels`, `unprotect_access_levels`

#### Attributes

* `access_level` - The access level allowed to perform the respective action (shows as 40 - "maintainer" if `user_id` or `group_id` are present).

* `access_level_description` - A description of the allowed access level(s), or the name of the user or group if `user_id` or `group_id` are present.

* `user_id` - If present, indicates that the user is allowed to perform the respective action.

* `group_id` - If present, indicates that the group is allowed to perform the respective action.
