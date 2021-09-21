# gitlab\_project

This resource allows you to create and manage projects within your GitLab group or within your user.

## Example Usage

```hcl
resource "gitlab_project" "example" {
  name        = "example"
  description = "My awesome codebase"

  visibility_level = "public"
}

# Project with custom push rules
resource "gitlab_project" "example-two" {
  name = "example-two"

  push_rules {
    author_email_regex     = "@example\\.com$"
    commit_committer_check = true
    member_check           = true
    prevent_secrets        = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the project.

* `path` - (Optional) The path of the repository.

* `namespace_id` - (Optional) The namespace (group or user) of the project. Defaults to your user.
  See [`gitlab_group`](group.html) for an example.

* `description` - (Optional) A description of the project.

* `tags` - (Optional) Tags (topics) of the project.

* `default_branch` - (Optional) The default branch for the project.

* `import_url` - (Optional) Git URL to a repository to be imported.

* `mirror` (Optional) Enables pull mirroring in a project. Default is `false`. For further information on mirroring,
consult the [gitlab documentation](https://docs.gitlab.com/ee/user/project/repository/repository_mirroring.html#repository-mirroring).

* `mirror_trigger_builds` (Optional) Pull mirroring triggers builds. Default is `false`.

* `mirror_overwrites_diverged_branches` (Optional) Pull mirror overwrites diverged branches.

* `only_mirror_protected_branches` (Optional) Only mirror protected branches.

* `request_access_enabled` - Allow users to request member access.

* `issues_enabled` - (Optional) Enable issue tracking for the project.

* `merge_requests_enabled` - (Optional) Enable merge requests for the project.

* `pipelines_enabled` - (Optional) Enable pipelines for the project.

* `approvals_before_merge` - (Optional) Number of merge request approvals required for merging. Default is 0.

* `wiki_enabled` - (Optional) Enable wiki for the project.

* `snippets_enabled` - (Optional) Enable snippets for the project.

* `container_registry_enabled` - (Optional) Enable container registry for the project.

* `lfs_enabled` - (Optional) Enable LFS for the project.

* `visibility_level` - (Optional) Set to `public` to create a public project.
  Valid values are `private`, `internal`, `public`.
  Repositories are created as private by default.

* `merge_method` - (Optional) Set to `ff` to create fast-forward merges
  Valid values are `merge`, `rebase_merge`, `ff`
  Repositories are created with `merge` by default

* `only_allow_merge_if_pipeline_succeeds` - (Optional) Set to true if you want allow merges only if a pipeline succeeds.

* `only_allow_merge_if_all_discussions_are_resolved` - (Optional) Set to true if you want allow merges only if all discussions are resolved.

* `allow_merge_on_skipped_pipeline` - (Optional) Set to true if you want to treat skipped pipelines as if they finished with success.

* `shared_runners_enabled` - (Optional) Enable shared runners for this project.

* `archived` - (Optional) Whether the project is in read-only mode (archived). Repositories can be archived/unarchived by toggling this parameter.

* `initialize_with_readme` - (Optional) Create main branch with first commit containing a README.md file.

* `packages_enabled` - (Optional) Enable packages repository for the project.

* `push_rules` (Optional) Push rules for the project (documented below).

* `template_name` - (Optional) When used without use_custom_template, name of a built-in project template. When used with use_custom_template, name of a custom project template. This option is mutually exclusive with `template_project_id`.

* `template_project_id` - (Optional)  When used with use_custom_template, project ID of a custom project template. This is preferable to using template_name since template_name may be ambiguous (enterprise edition). This option is mutually exclusive with `template_name`.

* `use_custom_template` - (Optional) Use either custom instance or group (with group_with_project_templates_id) project template (enterprise edition).

* `group_with_project_templates_id` - (Optional) For group-level custom templates, specifies ID of group from which all the custom project templates are sourced. Leave empty for instance-level templates. Requires use_custom_template to be true (enterprise edition).

* `pages_access_level` - (Optional) Enable pages access control
  Valid values are `disabled`, `private`, `enabled`, `public`.
  `private` is the default.

* `build_coverage_regex` - (Optional) Test coverage parsing for the project.

* `ci_config_path` - (Optional) Custom Path to CI config file.

## Attributes Reference

The following additional attributes are exported:

* `id` - Integer that uniquely identifies the project within the gitlab install.

* `path_with_namespace` - The path of the repository with namespace.

* `ssh_url_to_repo` - URL that can be provided to `git clone` to clone the
  repository via SSH.

* `http_url_to_repo` - URL that can be provided to `git clone` to clone the
  repository via HTTP.

* `web_url` - URL that can be used to find the project in a browser.

* `runners_token` - Registration token to use during runner setup.

* `remove_source_branch_after_merge` - Enable `Delete source branch` option by default for all new merge requests.

## Nested Blocks

### push_rules

For information on push rules, consult the [GitLab documentation](https://docs.gitlab.com/ce/push_rules/push_rules.html#push-rules).

#### Arguments

* `author_email_regex` - (Optional) All commit author emails must match this regex, e.g. `@my-company.com$`.

* `branch_name_regex` - (Optional) All branch names must match this regex, e.g. `(feature|hotfix)\/*`.

* `commit_message_regex` - (Optional) All commit messages must match this regex, e.g. `Fixed \d+\..*`.

* `commit_message_negative_regex` - (Optional) No commit message is allowed to match this regex, for example `ssh\:\/\/`.

* `file_name_regex` - (Optional) All commited filenames must not match this regex, e.g. `(jar|exe)$`.

* `commit_committer_check` - (Optional, bool) Users can only push commits to this repository that were committed with one of their own verified emails.

* `deny_delete_tag` - (Optional, bool) Deny deleting a tag.

* `member_check` - (Optional, bool) Restrict commits by author (email) to existing GitLab users.

* `prevent_secrets` - (Optional, bool) GitLab will reject any files that are likely to contain secrets.

* `reject_unsigned_commits` - (Optional, bool) Reject commit when it’s not signed through GPG.

* `max_file_size` - (Optional, int) Maximum file size (MB).

## Import

You can import a project state using `terraform import <resource> <id>`.  The
`id` can be whatever the [get single project api][get_single_project] takes for
its `:id` value, so for example:

```shell
$ terraform import gitlab_project.example richardc/example
```

[get_single_project]: https://docs.gitlab.com/ee/api/projects.html#get-single-project
[group_members_permissions]: https://docs.gitlab.com/ce/user/permissions.html#group-members-permissions
