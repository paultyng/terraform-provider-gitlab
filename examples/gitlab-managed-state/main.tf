terraform {
  backend "http" {
  }

  required_providers {
    gitlab = {
      source  = "gitlabhq/gitlab"
      version = "3.7.0"
    }
  }
}

provider "gitlab" {
  token = var.gitlab_token
}

resource "gitlab_group" "group" {
  name             = "My Group"
  path             = "my-group"
  description      = "Promoting Open Source Projects"
  visibility_level = "public"

  lifecycle {
    prevent_destroy = true
  }
}

resource "gitlab_group_membership" "group_membership" {
  for_each = var.group_members

  group_id     = gitlab_group.group.id
  user_id      = each.key
  access_level = each.value
}

resource "gitlab_project" "api" {
  name         = "api"
  description  = "An example project"
  namespace_id = gitlab_group.group.id

  only_allow_merge_if_all_discussions_are_resolved = true
  only_allow_merge_if_pipeline_succeeds            = true
  remove_source_branch_after_merge                 = true

  container_registry_enabled = false
  lfs_enabled                = false
  packages_enabled           = false
  request_access_enabled     = false
  shared_runners_enabled     = false
  snippets_enabled           = false
  wiki_enabled               = false

  tags = setunion(var.tags, ["api", "backend", "rest"])
}

resource "gitlab_branch_protection" "main" {
  project            = gitlab_project.api.id
  branch             = "main"
  push_access_level  = "developer"
  merge_access_level = "developer"
}

resource "gitlab_project_approval_rule" "default" {
  project            = gitlab_project.api.id
  name               = "Minimum one approval required"
  approvals_required = 1
}

resource "gitlab_project_level_mr_approvals" "default" {
  project_id                                 = gitlab_project.api.id
  merge_requests_author_approval             = false
  merge_requests_disable_committers_approval = true
  reset_approvals_on_push                    = true
}

resource "gitlab_pipeline_trigger" "default" {
  project     = gitlab_project.api.id
  description = "Used to trigger builds in bulk"
}

resource "gitlab_deploy_token" "default" {
  project = gitlab_project.api.id
  name    = "Default deploy token"
  scopes  = ["read_repository", "read_registry"]
}

resource "gitlab_label" "bug" {
  project     = gitlab_project.api.id
  name        = "bug"
  description = "issue for flagging bugs and/or errors"
  color       = "#ff0000"
}

resource "gitlab_label" "documentation" {
  project     = gitlab_project.api.id
  name        = "documentation"
  description = "issue for documentation updates"
  color       = "#00ff00"
}
