# gitlab\_group\_badge

This resource allows you to create and manage badges for your GitLab groups.
For further information, consult the [gitlab
documentation](https://docs.gitlab.com/ee/user/project/badges.html#group-badges).

## Example Usage

```hcl
resource "gitlab_group" "foo" {
  name = "foo-group"
}

resource "gitlab_group_badge" "example" {
  group   = gitlab_group.foo.id
  link_url  = "https://example.com/badge-123"
  image_url = "https://example.com/badge-123.svg"
}
```

## Argument Reference

The following arguments are supported:

* `group` - (Required) The id of the group to add the badge to.

* `link_url` - (Required) The url linked with the badge.

* `image_url` - (Required) The image url which will be presented on group overview.

## Attributes Reference

The resource exports the following attributes:

* `rendered_link_url` - The link_url argument rendered (in case of use of placeholders).

* `rendered_image_url` - The image_url argument rendered (in case of use of placeholders).

## Import

GitLab group badges can be imported using an id made up of `{group_id}:{badge_id}`,
 e.g.

```bash
terraform import gitlab_group_badge.foo 1:3
```
