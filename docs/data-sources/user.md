# gitlab\_user

Provide details about a specific user in the gitlab provider. Especially the ability to lookup the id for linking to other resources.

## Example Usage

```hcl
data "gitlab_user" "example" {
  username = "myuser"
}
```

### Example using `for_each`

```hcl
data "gitlab_user" "example-two" {
  for_each = toset(["user1", "user2", "user3"])

  username = each.value
}
```

## Argument Reference

The following arguments are supported:

* `email` - (Optional) The e-mail address of the user. (Requires administrator privileges)

* `username` - (Optional) The username of the user.

* `user_id` - (Optional) The ID of the user.

-> Only one of email, user_id, or username must be provided.

## Attributes Reference

* `id` - The unique id assigned to the user by the gitlab server.

* `username` - The username of the user.

* `email` - The e-mail address of the user.

* `name` - The name of the user.

* `is_admin` - Whether the user is an admin.

* `can_create_group` - Whether the user can create groups.

* `can_create_project` - Whether the user can create projects.

* `projects_limit` - Number of projects the user can create.

* `created_at` - Date the user was created at.

* `state` - Whether the user is active or blocked.

* `external` - Whether the user is external.

* `extern_uid` - The external UID of the user.

* `user_provider` - The UID provider of the user.

* `organization` - The organization of the user.

* `two_factor_enabled` - Whether user's two-factor auth is enabled.

* `avatar_url` - The avatar URL of the user.

* `bio` - The bio of the user.

* `location` - The location of the user.

* `skype` - Skype username of the user.

* `linkedin` - LinkedIn profile of the user.

* `twitter` - Twitter username of the user.

* `website_url` - User's website URL.

* `theme_id` - User's theme ID.

* `color_scheme_id` - User's color scheme ID.

* `last_sign_in_at` - Last user's sign-in date.

* `current_sign_in_at` - Current user's sign-in date.

-> Some attributes might not be returned depending on if you're an admin or not. Please refer to [Gitlab documentation](https://docs.gitlab.com/ce/api/users.html#single-user) for more details.
