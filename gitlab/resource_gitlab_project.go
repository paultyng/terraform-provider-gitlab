package gitlab

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	gitlab "github.com/xanzy/go-gitlab"
)

var resourceGitLabProjectSchema = map[string]*schema.Schema{
	"name": {
		Type:     schema.TypeString,
		Required: true,
	},
	"path": {
		Type:     schema.TypeString,
		Optional: true,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			if new == "" {
				return true
			}
			return old == new
		},
	},
	"path_with_namespace": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"namespace_id": {
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
	},
	"description": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"default_branch": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"import_url": {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: true,
	},
	"request_access_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	},
	"issues_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	},
	"merge_requests_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	},
	"pipelines_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	},
	"approvals_before_merge": {
		Type:     schema.TypeInt,
		Optional: true,
		Default:  0,
	},
	"wiki_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	},
	"snippets_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	},
	"container_registry_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	},
	"lfs_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	},
	"visibility_level": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice([]string{"private", "internal", "public"}, true),
		Default:      "private",
	},
	"merge_method": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice([]string{"merge", "rebase_merge", "ff"}, true),
		Default:      "merge",
	},
	"only_allow_merge_if_pipeline_succeeds": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
	},
	"only_allow_merge_if_all_discussions_are_resolved": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
	},
	"ssh_url_to_repo": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"http_url_to_repo": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"web_url": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"runners_token": {
		Type:      schema.TypeString,
		Computed:  true,
		Sensitive: true,
	},
	"shared_runners_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"tags": {
		Type:     schema.TypeSet,
		Optional: true,
		ForceNew: false,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Set:      schema.HashString,
	},
	"archived": {
		Type:        schema.TypeBool,
		Description: "Whether the project is archived.",
		Optional:    true,
		Default:     false,
	},
	"initialize_with_readme": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"squash_option": {
		Type:          schema.TypeString,
		Optional:      true,
		Default:  	   "default_off",
	},
	"remove_source_branch_after_merge": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"packages_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  true,
	},
	"push_rules": {
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"author_email_regex": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"branch_name_regex": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"commit_message_regex": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"commit_message_negative_regex": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"file_name_regex": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"commit_committer_check": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"deny_delete_tag": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"member_check": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"prevent_secrets": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"reject_unsigned_commits": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"max_file_size": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntAtLeast(0),
				},
			},
		},
	},
	"template_name": {
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"template_project_id"},
		ForceNew:      true,
	},
	"template_project_id": {
		Type:          schema.TypeInt,
		Optional:      true,
		ConflictsWith: []string{"template_name"},
		ForceNew:      true,
	},
	"use_custom_template": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"group_with_project_templates_id": {
		Type:     schema.TypeInt,
		Optional: true,
	},
	"pages_access_level": {
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "private",
		ValidateFunc: validation.StringInSlice([]string{"public", "private", "enabled", "disabled"}, true),
	},
	// The GitLab API requires that import_url is also set when mirror options are used
	// Ref: https://github.com/gitlabhq/terraform-provider-gitlab/pull/449#discussion_r549729230
	"mirror": {
		Type:         schema.TypeBool,
		Optional:     true,
		Default:      false,
		RequiredWith: []string{"import_url"},
	},
	"mirror_trigger_builds": {
		Type:         schema.TypeBool,
		Optional:     true,
		Default:      false,
		RequiredWith: []string{"import_url"},
	},
	"mirror_overwrites_diverged_branches": {
		Type:         schema.TypeBool,
		Optional:     true,
		Default:      false,
		RequiredWith: []string{"import_url"},
	},
	"only_mirror_protected_branches": {
		Type:         schema.TypeBool,
		Optional:     true,
		Default:      false,
		RequiredWith: []string{"import_url"},
	},
	"build_coverage_regex": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"ci_config_path": {
		Type:     schema.TypeString,
		Optional: true,
	},
}

func resourceGitlabProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabProjectCreate,
		Read:   resourceGitlabProjectRead,
		Update: resourceGitlabProjectUpdate,
		Delete: resourceGitlabProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: resourceGitLabProjectSchema,
	}
}

func resourceGitlabProjectSetToState(d *schema.ResourceData, project *gitlab.Project) error {
	d.SetId(fmt.Sprintf("%d", project.ID))
	d.Set("name", project.Name)
	d.Set("path", project.Path)
	d.Set("path_with_namespace", project.PathWithNamespace)
	d.Set("description", project.Description)
	d.Set("default_branch", project.DefaultBranch)
	d.Set("request_access_enabled", project.RequestAccessEnabled)
	d.Set("issues_enabled", project.IssuesEnabled)
	d.Set("merge_requests_enabled", project.MergeRequestsEnabled)
	d.Set("pipelines_enabled", project.JobsEnabled)
	d.Set("approvals_before_merge", project.ApprovalsBeforeMerge)
	d.Set("wiki_enabled", project.WikiEnabled)
	d.Set("snippets_enabled", project.SnippetsEnabled)
	d.Set("container_registry_enabled", project.ContainerRegistryEnabled)
	d.Set("lfs_enabled", project.LFSEnabled)
	d.Set("visibility_level", string(project.Visibility))
	d.Set("merge_method", string(project.MergeMethod))
	d.Set("only_allow_merge_if_pipeline_succeeds", project.OnlyAllowMergeIfPipelineSucceeds)
	d.Set("only_allow_merge_if_all_discussions_are_resolved", project.OnlyAllowMergeIfAllDiscussionsAreResolved)
	d.Set("namespace_id", project.Namespace.ID)
	d.Set("ssh_url_to_repo", project.SSHURLToRepo)
	d.Set("http_url_to_repo", project.HTTPURLToRepo)
	d.Set("web_url", project.WebURL)
	d.Set("runners_token", project.RunnersToken)
	d.Set("shared_runners_enabled", project.SharedRunnersEnabled)
	if err := d.Set("tags", project.TagList); err != nil {
		return err
	}
	d.Set("archived", project.Archived)
	d.Set("squash_option", project.SquashOption)
	d.Set("remove_source_branch_after_merge", project.RemoveSourceBranchAfterMerge)
	d.Set("packages_enabled", project.PackagesEnabled)
	d.Set("pages_access_level", string(project.PagesAccessLevel))
	d.Set("mirror", project.Mirror)
	d.Set("mirror_trigger_builds", project.MirrorTriggerBuilds)
	d.Set("mirror_overwrites_diverged_branches", project.MirrorOverwritesDivergedBranches)
	d.Set("only_mirror_protected_branches", project.OnlyMirrorProtectedBranches)
	d.Set("build_coverage_regex", project.BuildCoverageRegex)
	d.Set("ci_config_path", project.CIConfigPath)
	return nil
}

func resourceGitlabProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	options := &gitlab.CreateProjectOptions{
		Name:                             gitlab.String(d.Get("name").(string)),
		RequestAccessEnabled:             gitlab.Bool(d.Get("request_access_enabled").(bool)),
		IssuesEnabled:                    gitlab.Bool(d.Get("issues_enabled").(bool)),
		MergeRequestsEnabled:             gitlab.Bool(d.Get("merge_requests_enabled").(bool)),
		JobsEnabled:                      gitlab.Bool(d.Get("pipelines_enabled").(bool)),
		ApprovalsBeforeMerge:             gitlab.Int(d.Get("approvals_before_merge").(int)),
		WikiEnabled:                      gitlab.Bool(d.Get("wiki_enabled").(bool)),
		SnippetsEnabled:                  gitlab.Bool(d.Get("snippets_enabled").(bool)),
		ContainerRegistryEnabled:         gitlab.Bool(d.Get("container_registry_enabled").(bool)),
		LFSEnabled:                       gitlab.Bool(d.Get("lfs_enabled").(bool)),
		Visibility:                       stringToVisibilityLevel(d.Get("visibility_level").(string)),
		MergeMethod:                      stringToMergeMethod(d.Get("merge_method").(string)),
		OnlyAllowMergeIfPipelineSucceeds: gitlab.Bool(d.Get("only_allow_merge_if_pipeline_succeeds").(bool)),
		OnlyAllowMergeIfAllDiscussionsAreResolved: gitlab.Bool(d.Get("only_allow_merge_if_all_discussions_are_resolved").(bool)),
		SharedRunnersEnabled:                      gitlab.Bool(d.Get("shared_runners_enabled").(bool)),
		SquashOption:              				   gitlab.String(d.Get("squash_option").(string)),
		RemoveSourceBranchAfterMerge:              gitlab.Bool(d.Get("remove_source_branch_after_merge").(bool)),
		PackagesEnabled:                           gitlab.Bool(d.Get("packages_enabled").(bool)),
		Mirror:                                    gitlab.Bool(d.Get("mirror").(bool)),
		MirrorTriggerBuilds:                       gitlab.Bool(d.Get("mirror_trigger_builds").(bool)),
		BuildCoverageRegex:                        gitlab.String(d.Get("build_coverage_regex").(string)),
		CIConfigPath:                              gitlab.String(d.Get("ci_config_path").(string)),
	}

	if v, ok := d.GetOk("path"); ok {
		options.Path = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("namespace_id"); ok {
		options.NamespaceID = gitlab.Int(v.(int))
	}

	if v, ok := d.GetOk("description"); ok {
		options.Description = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("default_branch"); ok {
		options.DefaultBranch = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		options.TagList = stringSetToStringSlice(v.(*schema.Set))
	}

	if v, ok := d.GetOk("initialize_with_readme"); ok {
		options.InitializeWithReadme = gitlab.Bool(v.(bool))
	}

	if v, ok := d.GetOk("import_url"); ok {
		options.ImportURL = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("template_name"); ok {
		options.TemplateName = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("template_project_id"); ok {
		options.TemplateProjectID = gitlab.Int(v.(int))
	}

	if v, ok := d.GetOk("use_custom_template"); ok {
		options.UseCustomTemplate = gitlab.Bool(v.(bool))
	}

	if v, ok := d.GetOk("group_with_project_templates_id"); ok {
		options.GroupWithProjectTemplatesID = gitlab.Int(v.(int))
	}

	if v, ok := d.GetOk("pages_access_level"); ok {
		options.PagesAccessLevel = stringToAccessControlValue(v.(string))
	}

	if v, ok := d.GetOk("ci_config_path"); ok {
		options.CIConfigPath = gitlab.String(v.(string))
	}

	log.Printf("[DEBUG] create gitlab project %q", *options.Name)

	project, _, err := client.Projects.CreateProject(options)
	if err != nil {
		return err
	}

	// from this point onwards no matter how we return, resource creation
	// is committed to state since we set its ID
	d.SetId(fmt.Sprintf("%d", project.ID))

	// An import can be triggered by import_url or by creating the project from a template.
	if project.ImportStatus != "none" {
		log.Printf("[DEBUG] waiting for project %q import to finish", *options.Name)

		stateConf := &resource.StateChangeConf{
			Pending: []string{"scheduled", "started"},
			Target:  []string{"finished"},
			Timeout: 10 * time.Minute,
			Refresh: func() (interface{}, string, error) {
				status, _, err := client.ProjectImportExport.ImportStatus(d.Id())
				if err != nil {
					return nil, "", err
				}

				return status, status.ImportStatus, nil
			},
		}

		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf("error while waiting for project %q import to finish: %w", *options.Name, err)
		}

		// Read the project again, so that we can detect the default branch.
		project, _, err = client.Projects.GetProject(project.ID, nil)
		if err != nil {
			return fmt.Errorf("Failed to get project %q after completing import: %w", d.Id(), err)
		}
	}

	if d.Get("archived").(bool) {
		// strange as it may seem, this project is created in archived state...
		if _, _, err := client.Projects.ArchiveProject(d.Id()); err != nil {
			return fmt.Errorf("new project %q could not be archived: %w", d.Id(), err)
		}
	}

	if _, ok := d.GetOk("push_rules"); ok {
		err := editOrAddPushRules(client, d.Id(), d)
		var httpError *gitlab.ErrorResponse
		if errors.As(err, &httpError) && httpError.Response.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] Failed to edit push rules for project %q: %v", d.Id(), err)
			return errors.New("Project push rules are not supported in your version of GitLab")
		}
		if err != nil {
			return fmt.Errorf("Failed to edit push rules for project %q: %w", d.Id(), err)
		}
	}

	// default_branch cannot always be set during creation.
	// If the branch does not exist, the update will fail, so we also create it here.
	// See: https://gitlab.com/gitlab-org/gitlab/-/issues/333426
	// This logic may be removed when the above issue is resolved.
	if v, ok := d.GetOk("default_branch"); ok && project.DefaultBranch != "" && project.DefaultBranch != v.(string) {
		oldDefaultBranch := project.DefaultBranch
		newDefaultBranch := v.(string)

		log.Printf("[DEBUG] create branch %q for project %q", newDefaultBranch, d.Id())
		_, _, err := client.Branches.CreateBranch(project.ID, &gitlab.CreateBranchOptions{
			Branch: gitlab.String(newDefaultBranch),
			Ref:    gitlab.String(oldDefaultBranch),
		})
		if err != nil {
			return fmt.Errorf("Failed to create branch %q for project %q: %w", newDefaultBranch, d.Id(), err)
		}

		log.Printf("[DEBUG] set new default branch to %q for project %q", newDefaultBranch, d.Id())
		_, _, err = client.Projects.EditProject(project.ID, &gitlab.EditProjectOptions{
			DefaultBranch: gitlab.String(newDefaultBranch),
		})
		if err != nil {
			return fmt.Errorf("Failed to set default branch to %q for project %q: %w", newDefaultBranch, d.Id(), err)
		}

		log.Printf("[DEBUG] protect new default branch %q for project %q", newDefaultBranch, d.Id())
		_, _, err = client.ProtectedBranches.ProtectRepositoryBranches(project.ID, &gitlab.ProtectRepositoryBranchesOptions{
			Name: gitlab.String(newDefaultBranch),
		})
		if err != nil {
			return fmt.Errorf("Failed to protect default branch %q for project %q: %w", newDefaultBranch, d.Id(), err)
		}

		log.Printf("[DEBUG] unprotect old default branch %q for project %q", oldDefaultBranch, d.Id())
		_, err = client.ProtectedBranches.UnprotectRepositoryBranches(project.ID, oldDefaultBranch)
		if err != nil {
			return fmt.Errorf("Failed to unprotect undesired default branch %q for project %q: %w", oldDefaultBranch, d.Id(), err)
		}

		log.Printf("[DEBUG] delete old default branch %q for project %q", oldDefaultBranch, d.Id())
		_, err = client.Branches.DeleteBranch(project.ID, oldDefaultBranch)
		if err != nil {
			return fmt.Errorf("Failed to clean up undesired default branch %q for project %q: %w", oldDefaultBranch, d.Id(), err)
		}
	}

	var editProjectOptions gitlab.EditProjectOptions

	if v, ok := d.GetOk("mirror_overwrites_diverged_branches"); ok {
		editProjectOptions.MirrorOverwritesDivergedBranches = gitlab.Bool(v.(bool))
		editProjectOptions.ImportURL = gitlab.String(d.Get("import_url").(string))
	}

	if v, ok := d.GetOk("only_mirror_protected_branches"); ok {
		editProjectOptions.OnlyMirrorProtectedBranches = gitlab.Bool(v.(bool))
		editProjectOptions.ImportURL = gitlab.String(d.Get("import_url").(string))
	}

	if (editProjectOptions != gitlab.EditProjectOptions{}) {
		if _, _, err := client.Projects.EditProject(d.Id(), &editProjectOptions); err != nil {
			return fmt.Errorf("Could not update project %q: %w", d.Id(), err)
		}
	}

	return resourceGitlabProjectRead(d, meta)
}

func resourceGitlabProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] read gitlab project %s", d.Id())

	project, _, err := client.Projects.GetProject(d.Id(), nil)
	if err != nil {
		return err
	}
	if project.MarkedForDeletionAt != nil {
		log.Printf("[DEBUG] gitlab project %s is marked for deletion", d.Id())
		d.SetId("")
		return nil
	}

	if err := resourceGitlabProjectSetToState(d, project); err != nil {
		return err
	}

	log.Printf("[DEBUG] read gitlab project %q push rules", d.Id())

	pushRules, _, err := client.Projects.GetProjectPushRules(d.Id())
	var httpError *gitlab.ErrorResponse
	if errors.As(err, &httpError) && httpError.Response.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] Failed to get push rules for project %q: %v", d.Id(), err)
	} else if err != nil {
		return fmt.Errorf("Failed to get push rules for project %q: %w", d.Id(), err)
	}

	return d.Set("push_rules", flattenProjectPushRules(pushRules))
}

func resourceGitlabProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	options := &gitlab.EditProjectOptions{}
	transferOptions := &gitlab.TransferProjectOptions{}

	if d.HasChange("name") {
		options.Name = gitlab.String(d.Get("name").(string))
	}

	if d.HasChange("path") && (d.Get("path").(string) != "") {
		options.Path = gitlab.String(d.Get("path").(string))
	}

	if d.HasChange("namespace_id") {
		transferOptions.Namespace = gitlab.Int(d.Get("namespace_id").(int))
	}

	if d.HasChange("description") {
		options.Description = gitlab.String(d.Get("description").(string))
	}

	if d.HasChange("default_branch") {
		options.DefaultBranch = gitlab.String(d.Get("default_branch").(string))
	}

	if d.HasChange("visibility_level") {
		options.Visibility = stringToVisibilityLevel(d.Get("visibility_level").(string))
	}

	if d.HasChange("merge_method") {
		options.MergeMethod = stringToMergeMethod(d.Get("merge_method").(string))
	}

	if d.HasChange("only_allow_merge_if_pipeline_succeeds") {
		options.OnlyAllowMergeIfPipelineSucceeds = gitlab.Bool(d.Get("only_allow_merge_if_pipeline_succeeds").(bool))
	}

	if d.HasChange("only_allow_merge_if_all_discussions_are_resolved") {
		options.OnlyAllowMergeIfAllDiscussionsAreResolved = gitlab.Bool(d.Get("only_allow_merge_if_all_discussions_are_resolved").(bool))
	}

	if d.HasChange("request_access_enabled") {
		options.RequestAccessEnabled = gitlab.Bool(d.Get("request_access_enabled").(bool))
	}

	if d.HasChange("issues_enabled") {
		options.IssuesEnabled = gitlab.Bool(d.Get("issues_enabled").(bool))
	}

	if d.HasChange("merge_requests_enabled") {
		options.MergeRequestsEnabled = gitlab.Bool(d.Get("merge_requests_enabled").(bool))
	}

	if d.HasChange("pipelines_enabled") {
		options.JobsEnabled = gitlab.Bool(d.Get("pipelines_enabled").(bool))
	}

	if d.HasChange("approvals_before_merge") {
		options.ApprovalsBeforeMerge = gitlab.Int(d.Get("approvals_before_merge").(int))
	}

	if d.HasChange("wiki_enabled") {
		options.WikiEnabled = gitlab.Bool(d.Get("wiki_enabled").(bool))
	}

	if d.HasChange("snippets_enabled") {
		options.SnippetsEnabled = gitlab.Bool(d.Get("snippets_enabled").(bool))
	}

	if d.HasChange("shared_runners_enabled") {
		options.SharedRunnersEnabled = gitlab.Bool(d.Get("shared_runners_enabled").(bool))
	}

	if d.HasChange("tags") {
		options.TagList = stringSetToStringSlice(d.Get("tags").(*schema.Set))
	}

	if d.HasChange("container_registry_enabled") {
		options.ContainerRegistryEnabled = gitlab.Bool(d.Get("container_registry_enabled").(bool))
	}

	if d.HasChange("lfs_enabled") {
		options.LFSEnabled = gitlab.Bool(d.Get("lfs_enabled").(bool))
	}

	if d.HasChange("squash_option") {
		options.SquashOption = gitlab.String(d.Get("squash_option").(string))
	}

	if d.HasChange("remove_source_branch_after_merge") {
		options.RemoveSourceBranchAfterMerge = gitlab.Bool(d.Get("remove_source_branch_after_merge").(bool))
	}

	if d.HasChange("packages_enabled") {
		options.PackagesEnabled = gitlab.Bool(d.Get("packages_enabled").(bool))
	}

	if d.HasChange("pages_access_level") {
		options.PagesAccessLevel = stringToAccessControlValue(d.Get("pages_access_level").(string))
	}

	if d.HasChange("mirror") {
		options.ImportURL = gitlab.String(d.Get("import_url").(string))
		options.Mirror = gitlab.Bool(d.Get("mirror").(bool))
	}

	if d.HasChange("mirror_trigger_builds") {
		options.ImportURL = gitlab.String(d.Get("import_url").(string))
		options.MirrorTriggerBuilds = gitlab.Bool(d.Get("mirror_trigger_builds").(bool))
	}

	if d.HasChange("only_mirror_protected_branches") {
		options.ImportURL = gitlab.String(d.Get("import_url").(string))
		options.OnlyMirrorProtectedBranches = gitlab.Bool(d.Get("only_mirror_protected_branches").(bool))
	}

	if d.HasChange("mirror_overwrites_diverged_branches") {
		options.ImportURL = gitlab.String(d.Get("import_url").(string))
		options.MirrorOverwritesDivergedBranches = gitlab.Bool(d.Get("mirror_overwrites_diverged_branches").(bool))
	}

	if d.HasChange("build_coverage_regex") {
		options.BuildCoverageRegex = gitlab.String(d.Get("build_coverage_regex").(string))
	}

	if d.HasChange("ci_config_path") {
		options.CIConfigPath = gitlab.String(d.Get("ci_config_path").(string))
	}

	if *options != (gitlab.EditProjectOptions{}) {
		log.Printf("[DEBUG] update gitlab project %s", d.Id())
		_, _, err := client.Projects.EditProject(d.Id(), options)
		if err != nil {
			return err
		}
	}

	if *transferOptions != (gitlab.TransferProjectOptions{}) {
		log.Printf("[DEBUG] transferring project %s to namespace %d", d.Id(), transferOptions.Namespace)
		_, _, err := client.Projects.TransferProject(d.Id(), transferOptions)
		if err != nil {
			return err
		}
	}

	if d.HasChange("archived") {
		if d.Get("archived").(bool) {
			if _, _, err := client.Projects.ArchiveProject(d.Id()); err != nil {
				return fmt.Errorf("project %q could not be archived: %w", d.Id(), err)
			}
		} else {
			if _, _, err := client.Projects.UnarchiveProject(d.Id()); err != nil {
				return fmt.Errorf("project %q could not be unarchived: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("push_rules") {
		err := editOrAddPushRules(client, d.Id(), d)
		var httpError *gitlab.ErrorResponse
		if errors.As(err, &httpError) && httpError.Response.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] Failed to get push rules for project %q: %v", d.Id(), err)
			return errors.New("Project push rules are not supported in your version of GitLab")
		}
		if err != nil {
			return fmt.Errorf("Failed to edit push rules for project %q: %w", d.Id(), err)
		}
	}

	return resourceGitlabProjectRead(d, meta)
}

func resourceGitlabProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] Delete gitlab project %s", d.Id())

	_, err := client.Projects.DeleteProject(d.Id())
	if err != nil {
		return err
	}

	// Wait for the project to be deleted.
	// Deleting a project in gitlab is async.
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Target:  []string{"Deleted"},
		Refresh: func() (interface{}, string, error) {
			out, response, err := client.Projects.GetProject(d.Id(), nil)
			if err != nil {
				if response.StatusCode == 404 {
					return out, "Deleted", nil
				}
				log.Printf("[ERROR] Received error: %#v", err)
				return out, "Error", err
			}
			if out.MarkedForDeletionAt != nil {
				// Represents a Gitlab EE soft-delete
				return out, "Deleted", nil
			}
			return out, "Deleting", nil
		},

		Timeout:    10 * time.Minute,
		MinTimeout: 3 * time.Second,
		Delay:      5 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for project (%s) to become deleted: %s", d.Id(), err)
	}
	return nil
}

func editOrAddPushRules(client *gitlab.Client, projectID string, d *schema.ResourceData) error {
	log.Printf("[DEBUG] Editing push rules for project %q", projectID)

	editOptions := expandEditProjectPushRuleOptions(d)
	_, _, err := client.Projects.EditProjectPushRule(projectID, editOptions)
	if err == nil {
		return nil
	}

	var httpErr *gitlab.ErrorResponse
	if !errors.As(err, &httpErr) || httpErr.Response.StatusCode != http.StatusNotFound {
		return err
	}

	// A 404 could mean that the push rules need to be re-created.

	log.Printf("[DEBUG] Failed to edit push rules for project %q: %v", projectID, err)
	log.Printf("[DEBUG] Creating new push rules for project %q", projectID)

	addOptions := expandAddProjectPushRuleOptions(d)
	_, _, err = client.Projects.AddProjectPushRule(projectID, addOptions)
	if err != nil {
		return err
	}

	return nil
}

func expandEditProjectPushRuleOptions(d *schema.ResourceData) *gitlab.EditProjectPushRuleOptions {
	options := &gitlab.EditProjectPushRuleOptions{}

	if d.HasChange("push_rules.0.author_email_regex") {
		options.AuthorEmailRegex = gitlab.String(d.Get("push_rules.0.author_email_regex").(string))
	}

	if d.HasChange("push_rules.0.branch_name_regex") {
		options.BranchNameRegex = gitlab.String(d.Get("push_rules.0.branch_name_regex").(string))
	}

	if d.HasChange("push_rules.0.commit_message_regex") {
		options.CommitMessageRegex = gitlab.String(d.Get("push_rules.0.commit_message_regex").(string))
	}

	if d.HasChange("push_rules.0.commit_message_negative_regex") {
		options.CommitMessageNegativeRegex = gitlab.String(d.Get("push_rules.0.commit_message_negative_regex").(string))
	}

	if d.HasChange("push_rules.0.file_name_regex") {
		options.FileNameRegex = gitlab.String(d.Get("push_rules.0.file_name_regex").(string))
	}

	if d.HasChange("push_rules.0.commit_committer_check") {
		options.CommitCommitterCheck = gitlab.Bool(d.Get("push_rules.0.commit_committer_check").(bool))
	}

	if d.HasChange("push_rules.0.deny_delete_tag") {
		options.DenyDeleteTag = gitlab.Bool(d.Get("push_rules.0.deny_delete_tag").(bool))
	}

	if d.HasChange("push_rules.0.member_check") {
		options.MemberCheck = gitlab.Bool(d.Get("push_rules.0.member_check").(bool))
	}

	if d.HasChange("push_rules.0.prevent_secrets") {
		options.PreventSecrets = gitlab.Bool(d.Get("push_rules.0.prevent_secrets").(bool))
	}

	if d.HasChange("push_rules.0.reject_unsigned_commits") {
		options.RejectUnsignedCommits = gitlab.Bool(d.Get("push_rules.0.reject_unsigned_commits").(bool))
	}

	if d.HasChange("push_rules.0.max_file_size") {
		options.MaxFileSize = gitlab.Int(d.Get("push_rules.0.max_file_size").(int))
	}

	return options
}

func expandAddProjectPushRuleOptions(d *schema.ResourceData) *gitlab.AddProjectPushRuleOptions {
	options := &gitlab.AddProjectPushRuleOptions{}

	if v, ok := d.GetOk("push_rules.0.author_email_regex"); ok {
		options.AuthorEmailRegex = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("push_rules.0.branch_name_regex"); ok {
		options.BranchNameRegex = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("push_rules.0.commit_message_regex"); ok {
		options.CommitMessageRegex = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("push_rules.0.commit_message_negative_regex"); ok {
		options.CommitMessageNegativeRegex = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("push_rules.0.file_name_regex"); ok {
		options.FileNameRegex = gitlab.String(v.(string))
	}

	if v, ok := d.GetOk("push_rules.0.commit_committer_check"); ok {
		options.CommitCommitterCheck = gitlab.Bool(v.(bool))
	}

	if v, ok := d.GetOk("push_rules.0.deny_delete_tag"); ok {
		options.DenyDeleteTag = gitlab.Bool(v.(bool))
	}

	if v, ok := d.GetOk("push_rules.0.member_check"); ok {
		options.MemberCheck = gitlab.Bool(v.(bool))
	}

	if v, ok := d.GetOk("push_rules.0.prevent_secrets"); ok {
		options.PreventSecrets = gitlab.Bool(v.(bool))
	}

	if v, ok := d.GetOk("push_rules.0.reject_unsigned_commits"); ok {
		options.RejectUnsignedCommits = gitlab.Bool(v.(bool))
	}

	if v, ok := d.GetOk("push_rules.0.max_file_size"); ok {
		options.MaxFileSize = gitlab.Int(v.(int))
	}

	return options
}

func flattenProjectPushRules(pushRules *gitlab.ProjectPushRules) (values []map[string]interface{}) {
	if pushRules == nil {
		return []map[string]interface{}{}
	}

	return []map[string]interface{}{
		{
			"author_email_regex":            pushRules.AuthorEmailRegex,
			"branch_name_regex":             pushRules.BranchNameRegex,
			"commit_message_regex":          pushRules.CommitMessageRegex,
			"commit_message_negative_regex": pushRules.CommitMessageNegativeRegex,
			"file_name_regex":               pushRules.FileNameRegex,
			"commit_committer_check":        pushRules.CommitCommitterCheck,
			"deny_delete_tag":               pushRules.DenyDeleteTag,
			"member_check":                  pushRules.MemberCheck,
			"prevent_secrets":               pushRules.PreventSecrets,
			"reject_unsigned_commits":       pushRules.RejectUnsignedCommits,
			"max_file_size":                 pushRules.MaxFileSize,
		},
	}
}
