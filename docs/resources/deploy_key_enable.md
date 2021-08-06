# gitlab\_deploy\_key\_enable

This resource allows you to enable pre-existing deploy keys for your GitLab projects.

> **NOTE**: the GITLAB KEY_ID for the deploy key must be known

## Example Usage

```hcl
# A repo to host the deployment key
resource "gitlab_project" "parent" {
  name = "parent_project"
}

# A second repo to use the deployment key from the parent project
resource "gitlab_project" "foo" {
  name = "foo_project"
}

# Upload a deployment key for the parent repo
resource "gitlab_deploy_key" "parent" {
  project = "${gitlab_project.parent.id}"
  title = "Example deploy key"
  key = "ssh-rsa AAAA..."
}

# Enable the deployment key on the second repo
resource "gitlab_deploy_key_enable" "foo" {
  project = "${gitlab_project.foo.id}"
  key_id = "${gitlab_deploy_key.parent.id}"
}
```

## Argument Reference

The following arguments are supported:

* `project` - (Required, string) The name or id of the project to add the deploy key to.

* `key_id` - (Required, string) The Gitlab key id for the pre-existing deploy key

* `can_push` - (Optional, boolean) Allow the deploy key to be used to push changes to the project. Defaults to `false`.

## Import

GitLab enabled deploy keys can be imported using an id made up of `{project_id}:{deploy_key_id}`, e.g.

```
$ terraform import gitlab_deploy_key_enable.example 12345:67890
```
