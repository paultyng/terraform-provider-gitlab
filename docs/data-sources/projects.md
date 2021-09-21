# gitlab\_projects

Provide details about a list of projects in the Gitlab provider. Listing all projects and group projects with [project filtering](https://docs.gitlab.com/ee/api/projects.html#list-user-projects) or [group project filtering](https://docs.gitlab.com/ee/api/groups.html#list-a-groups-projects) is supported.

> **NOTE**: This data source supports all available filters exposed by the `xanzy/go-gitlab` package, which might not expose all available filters exposed by the Gitlab APIs.

## Example Usage

### List projects within a group tree

```hcl
data "gitlab_group" "mygroup" {
  full_path = "mygroup"
}

data "gitlab_projects" "group_projects" {
  group_id          = data.gitlab_group.mygroup.id
  order_by          = "name"
  include_subgroups = true
  with_shared       = false
}
```

### List projects using the search syntax

```hcl
data "gitlab_projects" "projects" {
  search              = "postgresql"
  visibility          = "private"
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Optional) The ID of the group owned by the authenticated user to look projects for within. Cannot be used with `min_access_level`, `with_programming_language` or `statistics`.

* `per_page`: The maximum number of projects to return in one paginated API call, limited to `100`. Default is `20`.

* `max_queryable_pages` Prevents overloading your Gitlab instance in case of a misconfiguration. Default is `10`.

* `archived` - (Optional) Limit by archived status.

* `visibility` - (Optional) Limit by visibility `public`, `internal`, or `private`.

* `order_by` - (Optional) Return projects ordered by `id`, `name`, `path`, `created_at`, `updated_at`, or `last_activity_at` fields. Default is `created_at`.

* `sort` - (Optional) Return projects sorted in `asc` or `desc` order. Default is `desc`.

* `search` - (Optional) Return list of authorized projects matching the search criteria.

* `simple` - (Optional) Return only the ID, URL, name, and path of each project.

* `owned` - (Optional) Limit by projects owned by the current user.

* `starred` - (Optional) Limit by projects starred by the current user.

* `with_issues_enabled` - (Optional) Limit by projects with issues feature enabled. Default is `false`.

* `with_merge_requests_enabled` - (Optional) Limit by projects with merge requests feature enabled. Default is `false`.

* `with_shared` - (Optional) Include projects shared to this group. Default is `true`. Needs `group_id`.

* `include_subgroups` - (Optional) Include projects in subgroups of this group. Default is `false`. Needs `group_id`.

* `min_access_level` - (Optional) Limit to projects where current user has at least this access level, refer to the [official documentation](https://docs.gitlab.com/ee/api/members.html) for values. Cannot be used with `group_id`.

* `with_custom_attributes` - (Optional) Include custom attributes in response _(admins only)_.

* `membership` - (Optional) Limit by projects that the current user is a member of.

* `statistics` - (Optional) Include project statistics. Cannot be used with `group_id`.

* `with_programming_language` - (Optional) Limit by projects which use the given programming language. Cannot be used with `group_id`.

## Attributes Reference

The following attributes are exported:

* `projects` - A list containing the projects matching the supplied arguments

Projects items have the following fields:

* `id` - The ID of the project.

* `name` - The name of the project.

* `description`

* `default_branch`

* `public` - Whether the project is public.

* `visibility` - The visibility of the project.

* `ssh_url_to_repo` - The SSH clone URL of the project.

* `http_url_to_repo` - The HTTP clone URL of the project.

* `web_url`

* `readme_url`

* `tag_list` - A set of the project topics (formerly called "project tags").

* `owner` - The owner of the project, due to Terraform aggregate types limitations, this field's attributes are accessed with the `owner.0` prefix. Structure is documented below.

* `name_with_namespace` - In `group / subgroup / project` or `user / project` format.

* `path`

* `path_with_namespace` - In `group/subgroup/project` or `user/project` format.

* `issues_enabled`

* `open_issues_count`

* `merge_requests_enabled`

* `approvals_before_merge` - The numbers of approvals needed in a merge requests.

* `jobs_enabled` - Whether pipelines are enabled for the project.

* `wiki_enabled`

* `snippets_enabled`

* `resolve_outdated_diff_discussions`

* `container_registry_enabled`

* `created_at`

* `last_activity_at`

* `creator_id`

* `namespace` namespace of the project, due to Terraform aggregate types limitations, this field's attributes are accessed with the `namespace.0` prefix. Structure is documented below.

* `import_status`

* `import_error`

* `permissions` permissions for the project, due to Terraform aggregate types limitations, this field's attributes are accessed with the `permissions.0` prefix. Structure is documented below.

* `archived`

* `avatar_url`

* `shared_runners_enabled`

* `forks_count`

* `star_count`

* `runners_token`

* `public_builds`

* `only_allow_merge_if_pipeline_succeeds`

* `only_allow_merge_if_all_discussions_are_resolved`

* `allow_merge_on_skipped_pipeline`

* `lfs_enabled`

* `request_access_enabled`

* `merge_method`

* `forked_from_project` If this project has been forked from another project, due to Terraform aggregate types limitations, this field's attributes are accessed with the `forked_from_project.0` prefix. Structure is documented below.

* `mirror`

* `mirror_user_id`

* `mirror_trigger_builds`

* `only_mirror_protected_branches`

* `mirror_overwrites_diverged_branches`

* `shared_with_groups` List of groups with which the project is shared, the structure is documented below.

* `statistics`

* `_links`

* `ci_config_path`

* `custom_attributes`

* `build_coverage_regex`

The `owner` attribute exposes the following sub-attributes:

> **NOTE**: These sub-attributes are only populated if the Gitlab token used has an administrator scope.

* `id`

* `username`

* `name`

* `state`

* `avatar_url`

* `website_url`

The `namespace` attribute exposes the following sub-attributes:

* `id`

* `name`

* `path`

* `kind` Whether the namespace is a `group` or a `user`.

* `full_path`

The `permissions` attribute exposes the following sub-attributes:

* `project_access` Structure is documented below.

* `group_access` Structure is documented below.

The `permissions.0.project_access` attribute exposes the following sub-attributes:

* `access_level`

* `notification_level`

The `permissions.0.group_access` attribute exposes the following sub-attributes:

* `access_level`

* `notification_level`

The `forked_from_project` attribute exposes the following sub-attributes:

* `id`

* `http_url_to_repo`

* `name`

* `name_with_namespace`

* `path`

* `path_with_namespace`

* `web_url`

The `shared_with_groups` list objects expose the following attributes:

* `group_id`

* `group_access_level`

* `group_name`
