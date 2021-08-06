# gitlab\_branch\_protection

This resource allows you to protect a specific branch by an access level so that the user with less access level cannot Merge/Push to the branch.

-> The `allowed_to_push`, `allowed_to_merge` and `code_owner_approval_required` arguments require a GitLab Premium account or above.  Please refer to [Gitlab API documentation](https://docs.gitlab.com/ee/api/protected_branches.html) for further information.

## Example Usage

```hcl
resource "gitlab_branch_protection" "BranchProtect" {
  project                      = "12345"
  branch                       = "BranchProtected"
  push_access_level            = "developer"
  merge_access_level           = "developer"
  code_owner_approval_required = true
  allowed_to_push {
    user_id = 5
  }
  allowed_to_push {
    user_id = 521
  }
  allowed_to_merge {
    user_id = 15
  }
  allowed_to_merge {
    user_id = 37
  }
}
```

### Example using dynamic block

```hcl
resource "gitlab_branch_protection" "main" {
  project                      = "12345"
  branch                       = "main"
  push_access_level            = "maintainer"
  merge_access_level           = "maintainer"

  dynamic "allowed_to_push" {
    for_each = [50, 55, 60]
    content {
      user_id = allowed_to_push.value
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `project` - (Required) The id of the project.

* `branch` - (Required) Name of the branch.

* `push_access_level` - (Optional) One of five levels of access to the project. Valid values are: `no one`, `developer`, `maintainer`, `admin`.

* `merge_access_level` - (Optional) One of five levels of access to the project. Valid values are: `no one`, `developer`, `maintainer`, `admin`.

* `allowed_to_push`, `allowed_to_merge` - (Optional) One or more `allowed_to_push`, `allowed_to_merge` blocks as defined below.

* `code_owner_approval_required` (Optional) Bool, defaults to false. Can be set to true to require code owner approval before merging.

---

An `allowed_to_push` or `allowed_to_merge` block supports the following arguments:

* `user_id` - (Required) The ID of a GitLab user allowed to perform the relevant action. Mutually exclusive with `group_id`.

* `group_id` - (Required) The ID of a GitLab group allowed to perform the relevant action. Mutually exclusive with `user_id`.

## Attributes Reference

The following attributes are exported:

* `branch_protection_id` - The ID of the branch protection (not the branch name).

* The `allowed_to_push` and `allowed_to_merge` blocks export the `access_level_description` field, which contains a textual description of the access level, user or group allowed to perform the relevant action.

## Import

Gitlab protected branches can be imported with a key composed of `<project_id>:<branch>`, e.g.

```
$ terraform import gitlab_branch_protection.BranchProtect "12345:main"
```
