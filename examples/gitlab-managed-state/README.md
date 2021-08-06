# Terraform provider example

## Getting started

### Gitlab access token

First create an access token, with the `api` scope and save that in a secure place (you won't be able to see it again). See the [documentation](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) for further information.
The simplest way to provide the token is in a variable file:

```shell
$ ACCESS_TOKEN="YOUR_GITLAB_TOKEN"
$ echo gitlab_token = "\"${ACCESS_TOKEN}\"" >> terraform.tfvars
```

### Kick-starting the backend

Then you need to initialize the state file, by simply replacing the variables and running the following command:

```shell
$ USERNAME="YOUR_GITLAB_USERNAME"
$ ACCESS_TOKEN="YOUR_GITLAB_TOKEN"
$ PROJECT_ID="12345678"
$ STATE_NAME="default"

$ ADDRESS="https://gitlab.com/api/v4/projects/${PROJECT_ID}/terraform/state/${STATE_NAME}"

$ terraform init \
    -backend-config="address=${ADDRESS}" \
    -backend-config="lock_address=${ADDRESS}/lock" \
    -backend-config="unlock_address=${ADDRESS}/lock" \
    -backend-config="username=${USERNAME}" \
    -backend-config="password=${ACCESS_TOKEN}" \
    -backend-config="lock_method=POST" \
    -backend-config="unlock_method=DELETE" \
    -backend-config="retry_wait_min=5"
    -backend-config="password=${ACCESS_TOKEN}"
```

The available arguments and valid values can be seen in the [Terraform documentation](https://www.terraform.io/docs/language/settings/backends/http.html#configuration-variables).

Check the newly created state by running `terraform plan`.

## Creating the infrastructure

Once the backend is initiated, you can start to create your resources:

```shell
$ terraform apply
```

## Destroying all resources

Similarly, all resources can be deleted or scheduled for deletion by running:

```shell
$ terraform destroy
```

## References

1. [GitLab managed Terraform State](https://docs.gitlab.com/ee/user/infrastructure/terraform_state.html)
2. [Infrastructure as code with Terraform and GitLab](https://docs.gitlab.com/ee/user/infrastructure/index.html)
