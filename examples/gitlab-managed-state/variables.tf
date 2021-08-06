variable "gitlab_token" {
  type        = string
  description = "GitLab personal access token"
  sensitive   = true
}

variable "group_members" {
  description = <<EOF
All members of the group and its [access level](https://docs.gitlab.com/ee/user/permissions.html#project-members-permissions).
Possible values are: `guest`, `reporter`, `developer`, `maintainer`, `owner`
EOF
  type        = map(string)
  default = {
    "1234567" = "owner"
    "2345678" = "developer"
  }
}

variable "tags" {
  type        = list(string)
  description = "A list of tags (topics) of the project"
  default     = ["gitlab", "terraform"]
}
