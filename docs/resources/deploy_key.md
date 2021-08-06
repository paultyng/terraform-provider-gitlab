# gitlab\_deploy\_key

This resource allows you to create and manage [deploy keys](https://docs.gitlab.com/ee/user/project/deploy_keys/) for your GitLab projects.

## Example Usage

```hcl
resource "gitlab_deploy_key" "example" {
  project = "example/deploying"
  title   = "Example deploy key"
  key     = "ssh-rsa AAAA..."
}
```

## Argument Reference

The following arguments are supported:

* `project` - (Required, string) The name or id of the project to add the deploy key to.

* `title` - (Required, string) A title to describe the deploy key with.

* `key` - (Required, string) The public ssh key body.

* `can_push` - (Optional, boolean) Allow this deploy key to be used to push changes to the project. Defaults to `false`.

## Import

GitLab deploy keys can be imported using an id made up of `{project_id}:{deploy_key_id}`, e.g.

```
$ terraform import gitlab_deploy_key.test 1:3
```
