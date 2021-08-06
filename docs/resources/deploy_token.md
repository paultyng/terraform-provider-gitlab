# gitlab\_deploy\_token

This resource allows you to create and manage deploy token for your GitLab projects and groups. Please refer to [Gitlab documentation](https://docs.gitlab.com/ee/user/project/deploy_tokens/) for further information.

## Example Usage - Project

```hcl
resource "gitlab_deploy_token" "example" {
  project    = "example/deploying"
  name       = "Example deploy token"
  username   = "example-username"
  expires_at = "2020-03-14T00:00:00.000Z"
  
  scopes = [ "read_repository", "read_registry" ]
}

resource "gitlab_deploy_token" "example-two" {
  project    = "12345678"
  name       = "Example deploy token expires in 24h"
  expires_at = timeadd(timestamp(), "24h")
}
```

## Example Usage - Group

```hcl
resource "gitlab_deploy_token" "example" {
  group      = "example/deploying"
  name       = "Example group deploy token"

  scopes = [ "read_repository" ]
}
```

## Argument Reference

The following arguments are supported:

* `project` - (Required, string) The name or id of the project to add the deploy token to.
  Either `project` or `group` must be set.

* `group` - (Required, string) The name or id of the group to add the deploy token to.
  Either `project` or `group` must be set.

* `name` - (Required, string) A name to describe the deploy token with.

* `username` - (Optional, string) A username for the deploy token. Default is `gitlab+deploy-token-{n}`.

* `expires_at` - (Optional, string) Time the token will expire it, RFC3339 format. Will not expire per default.

* `scopes` - (Required, set of strings) Valid values: `read_repository`, `read_registry`.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `token` - The secret token. This is only populated when creating a new deploy token.
