# gitlab_group_variable

This resource allows you to create and manage CI/CD variables for your GitLab groups.
For further information on variables, consult the [gitlab
documentation](https://docs.gitlab.com/ce/ci/variables/README.html#variables).

## Example Usage

```hcl
resource "gitlab_group_variable" "example" {
   group             = "12345"
   key               = "group_variable_key"
   value             = "group_variable_value"
   protected         = false
   masked            = false
   environment_scope = "example"
}
```

## Argument Reference

The following arguments are supported:

- `group` - (Required, string) The name or id of the group to add the hook to.

- `key` - (Required, string) The name of the variable.

- `value` - (Required, string) The value of the variable.

- `variable_type` - (Optional, string) The type of a variable. Available types are: env_var (default) and file.

- `protected` - (Optional, boolean) If set to `true`, the variable will be passed only to pipelines running on protected branches and tags. Defaults to `false`.

- `masked` - (Optional, boolean) If set to `true`, the value of the variable will be hidden in job logs. The value must meet the [masking requirements](https://docs.gitlab.com/ee/ci/variables/#masked-variables). Defaults to `false`.

- `environment_scope` - (Optional, string) Set to limit the environment scope of a CI/CD variable by defining which environments it can be available for. Only available with Premium license and above, from [GitLab 13.11](https://docs.gitlab.com/ee/ci/environments/index.html#scoping-environments-with-specs).

## Import

GitLab group variables can be imported using an id made up of `groupid:variablename`, e.g.

```
$ terraform import gitlab_group_variable.example 12345:group_variable_key
```
