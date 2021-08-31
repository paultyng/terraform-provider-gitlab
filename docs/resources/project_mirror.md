# gitlab\_project_mirror

This resource allows you to add a mirror target for the repository, all changes will be synced to the remote target.

-> This is for *pushing* changes to a remote repository. *Pull Mirroring* can be configured using a combination of the
`import_url`, `mirror`, and `mirror_trigger_builds` properties on the `gitlab_project` resource.

For further information on mirroring, consult the
[gitlab documentation](https://docs.gitlab.com/ee/user/project/repository/repository_mirroring.html#repository-mirroring).

## Example Usage

```hcl
resource "gitlab_project_mirror" "foo" {
  project = "1"
  url = "https://username:password@github.com/org/repository.git"
}
```

## Argument Reference

The following arguments are supported:

* `project` - (Required) The id of the project.

* `url` - (Required) The URL of the remote repository to be mirrored.

~> Due to limitations of the GitLab API, the provider cannot update the
username and passwords that are present in the URL. To change the username or
the password, we recommend you to [replace](https://www.terraform.io/docs/cli/commands/plan.html#replace-address) the resources.

* `enabled` - Determines if the mirror is enabled.

* `only_protected_branches` - Determines if only protected branches are mirrored.

* `keep_divergent_refs` - Determines if divergent refs are skipped.

## Import

GitLab project mirror can be imported using an id made up of `project_id:mirror_id`, e.g.

```
$ terraform import gitlab_project_mirror.foo "12345:1337"
```
