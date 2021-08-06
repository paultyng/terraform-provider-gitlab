# gitlab\_project\_environment

This resource allows you to create and manage environments for your GitLab
projects. For further information on project environments, consult the [GitLab
documentation](https://docs.gitlab.com/ee/ci/environments/) and the [API
documentation](https://docs.gitlab.com/ee/api/environments.html).

## Example Usage

```hcl
resource "gitlab_project_environment" "example-one" {
  project      = 5
  name         = "Production/web"
  external_url = "https://example.com"
}
```

## Argument Reference

The following arguments are supported:

* `project` - (Required, string) The name or id of the project to add the project environments.

* `name` - (Required) The name of the project environment.

* `external_url` - (Optional) External URL for this project environment.

## Attributes Reference

---

The following attributes are exported:

* `state` - The state of the environment (e.g. `available`).

## Import

GitLab project environments can be imported using an id consisting of `project-id:environment-id`, e.g.

```
$ terraform import gitlab_project_environment.example "12345:6"
```
