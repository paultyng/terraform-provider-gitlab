# gitlab\_group

This resource allows you to create and manage GitLab groups.
Note your provider will need to be configured with admin-level access for this resource to work.

## Example Usage

```hcl
resource "gitlab_group" "example" {
  name        = "example"
  path        = "example"
  description = "An example group"
}

# Create a project in the example group
resource "gitlab_project" "example" {
  name         = "example"
  description  = "An example project"
  namespace_id = gitlab_group.example.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of this group.

* `path` - (Required) The path of the group.

* `description` - (Optional) The description of the group.

* `lfs_enabled` - (Optional) Boolean, defaults to true.  Whether to enable LFS
support for projects in this group.

* `request_access_enabled` - (Optional) Boolean, defaults to false.  Whether to
enable users to request access to the group.

* `visibility_level` - (Optional) The group's visibility. Can be `private`, `internal`, or `public`.

* `share_with_group_lock` - (Optional) Boolean, defaults to false.  Prevent sharing
a project with another group within this group.

* `project_creation_level` - (Optional), defaults to Maintainer.
Determine if developers can create projects
in the group. Can be noone (No one), maintainer (Maintainers),
or developer (Developers + Maintainers).

* `auto_devops_enabled` - (Optional) Boolean, defaults to false.  Default to Auto
DevOps pipeline for all projects within this group.

* `emails_disabled` - (Optional) Boolean, defaults to false.  Disable email notifications

* `mentions_disabled` - (Optional) Boolean, defaults to false.  Disable the capability
of a group from getting mentioned

* `subgroup_creation_level` - (Optional), defaults to Owner.
 Allowed to create subgroups.
Can be owner (Owners), or maintainer (Maintainers).

* `require_two_factor_authentication` - (Optional) Boolean, defaults to false.
equire all users in this group to setup Two-factor authentication.

* `two_factor_grace_period` - (Optional) Int, defaults to 48.
Time before Two-factor authentication is enforced (in hours).

* `parent_id` - (Optional) Integer, id of the parent group (creates a nested group).

## Attributes Reference

The resource exports the following attributes:

* `id` - The unique id assigned to the group by the GitLab server.  Serves as a
  namespace id where one is needed.
  
* `full_path` - The full path of the group.

* `full_name` - The full name of the group.

* `web_url` - Web URL of the group.

* `runners_token` - The group level registration token to use during runner setup.

* `default_branch_protection` - (Optional) Int, defaults to 2.
Whether developers and maintainers can push to the applicable default branch.
0 no protection, 1 partial protection, 2 full protection
https://docs.gitlab.com/ee/api/groups.html#options-for-default_branch_protection

## Import

You can import a group state using `terraform import <resource> <id>`.  The
`id` can be whatever the [details of a group][details_of_a_group] api takes for
its `:id` value, so for example:

```shell
$ terraform import gitlab_group.example example
```

[details_of_a_group]: https://docs.gitlab.com/ee/api/groups.html#details-of-a-group
