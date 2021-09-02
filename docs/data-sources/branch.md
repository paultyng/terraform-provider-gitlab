# gitlab\_branch

Provides details about a specific branch in the gitlab provider.

## Example Usage

### Branch name and url encoded project path

```hcl
data "gitlab_branch" "example" {
  name = "branch-name"
  project = "namespace/project-name"
}
```

### Branch name and test project ID

```hcl
data "gitlab_branch" "example" {
  name = "branch-name"
  project = "270771234"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the branch.

* `project` - (Required) The full path of the group.

## Attributes Reference

The resource exports the following attributes:

* `web_url` - The url of the created branch (https)
* `default` - Bool, true if branch is the default branch for the project
* `can_push` - Bool, true if you can push to the branch
* `merged` - Bool, true if the branch has been merged into it's parent
* `commit` - The list of group members.
  * `id` - The unique id assigned to the commit by gitlab.
  * `short_id` - The short id assigned to the commit by gitlab
  * `author_email` - The email of the author.
  * `author_name` - The name of the author.
  * `authored_date` - The date which the commit was authored
  * `commited_date` - The date at which the commit was pushed.
  * `committer_email` - The email of the user that committed.
  * `commiter_name` - The name of the user that committed.
  * `title` - The title of the commit
  * `message` - The commit message
  * `parent_ids` - The id of the parents of the commit
