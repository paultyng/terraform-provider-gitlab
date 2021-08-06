# gitlab\_project\_level\_mr\_approvals

This resource allows you to configure project-level MR approvals. for your GitLab projects.
For further information on merge request approvals, consult the [GitLab API documentation](https://docs.gitlab.com/ee/api/merge_request_approvals.html#project-level-mr-approvals).

## Example Usage

```hcl
resource "gitlab_project" "foo" {
  name              = "Example"
  description       = "My example project"
}

resource "gitlab_project_level_mr_approvals" "foo" {
  project_id                                     = gitlab_project.foo.id
  reset_approvals_on_push                        = true
  disable_overriding_approvers_per_merge_request = false
  merge_requests_author_approval                 = false
  merge_requests_disable_committers_approval     = true
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) The ID of the project to change MR approval configuration.

* `reset_approvals_on_push` - (Optional) Set to `true` if you want to remove all approvals in a merge request when new commits are pushed to its source branch. Default is `true`.

* `disable_overriding_approvers_per_merge_request` - (Optional) By default, users are able to edit the approval rules in merge requests. If set to true,
the approval rules for all new merge requests will be determined by the default approval rules. Default is `false`.

* `merge_requests_author_approval` - (Optional) Set to `true` if you want to allow merge request authors to self-approve merge requests. Authors
also need to be included in the approvers list in order to be able to approve their merge request. Default is `false`.

* `merge_requests_disable_committers_approval` - (Optional) Set to `true` if you want to prevent approval of merge requests by merge request committers. Default is `false`.

## Import

You can import an approval configuration state using `terraform import <resource> <project_id>:<approval_rule_id>`.

For example:

```
$ terraform import gitlab_project_level_mr_approvals.foo 1234:53
```
