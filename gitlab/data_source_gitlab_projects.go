// lintignore: S031 // TODO: Resolve this tfproviderlint issue

package gitlab

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/mitchellh/hashstructure"
	"github.com/xanzy/go-gitlab"
)

// Schemas

// WARN: go-gitlab may not be up-to-date with Gitlab exposed options
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-projects
// https://docs.gitlab.com/ee/api/projects.html#list-all-projects

// Helper functions
func flattenProjectPermissions(permissions *gitlab.Permissions) []map[string]interface{} {
	m := make(map[string]interface{}, 2)
	if permissions != nil {
		if permissions.ProjectAccess != nil {
			m["project_access"] = map[string]int{
				"access_level":       int(permissions.ProjectAccess.AccessLevel),
				"notification_level": int(permissions.ProjectAccess.NotificationLevel),
			}
		}
		if permissions.GroupAccess != nil {
			m["group_access"] = map[string]int{
				"access_level":       int(permissions.GroupAccess.AccessLevel),
				"notification_level": int(permissions.GroupAccess.NotificationLevel),
			}
		}
	}
	return []map[string]interface{}{m}
}

func flattenProjectNamespace(namespace *gitlab.ProjectNamespace) (values []map[string]interface{}) {
	if namespace != nil {
		values = []map[string]interface{}{
			{
				"id":        namespace.ID,
				"name":      namespace.Name,
				"path":      namespace.Path,
				"kind":      namespace.Kind,
				"full_path": namespace.FullPath,
			},
		}
	}
	return values
}

func flattenProjectLinks(links *gitlab.Links) (values map[string]string) {
	if links != nil {
		values = map[string]string{
			"self":           links.Self,
			"issues":         links.Issues,
			"merge_requests": links.MergeRequests,
			"repo_branches":  links.RepoBranches,
			"labels":         links.Labels,
			"events":         links.Events,
			"members":        links.Members,
		}
	}
	return values
}

func flattenForkedFromProject(forked *gitlab.ForkParent) (values []map[string]interface{}) {
	if forked != nil {
		values = []map[string]interface{}{
			{
				"http_url_to_repo":    forked.HTTPURLToRepo,
				"id":                  forked.ID,
				"name":                forked.Name,
				"name_with_namespace": forked.NameWithNamespace,
				"path":                forked.Path,
				"path_with_namespace": forked.PathWithNamespace,
				"web_url":             forked.WebURL,
			},
		}
	}
	return values
}

func flattenGitlabBasicUser(user *gitlab.User) (values []map[string]interface{}) {
	if user != nil {
		values = []map[string]interface{}{
			{
				"id":          user.ID,
				"username":    user.Username,
				"name":        user.Name,
				"state":       user.State,
				"avatar_url":  user.AvatarURL,
				"website_url": user.WebsiteURL,
			},
		}
	}
	return values
}

func flattenProjects(projects []*gitlab.Project) (values []map[string]interface{}) {
	if projects != nil { // nolint // TODO: Resolve this golangci-lint issue: S1031: unnecessary nil check around range (gosimple)
		for _, project := range projects {
			v := map[string]interface{}{
				"id":                                    project.ID,
				"description":                           project.Description,
				"default_branch":                        project.DefaultBranch,
				"public":                                project.Public,
				"visibility":                            string(project.Visibility),
				"ssh_url_to_repo":                       project.SSHURLToRepo,
				"http_url_to_repo":                      project.HTTPURLToRepo,
				"web_url":                               project.WebURL,
				"readme_url":                            project.ReadmeURL,
				"tag_list":                              project.TagList,
				"owner":                                 flattenGitlabBasicUser(project.Owner),
				"name":                                  project.Name,
				"name_with_namespace":                   project.NameWithNamespace,
				"path":                                  project.Path,
				"path_with_namespace":                   project.PathWithNamespace,
				"issues_enabled":                        project.IssuesEnabled,
				"open_issues_count":                     project.OpenIssuesCount,
				"merge_requests_enabled":                project.MergeRequestsEnabled,
				"approvals_before_merge":                project.ApprovalsBeforeMerge,
				"jobs_enabled":                          project.JobsEnabled,
				"wiki_enabled":                          project.WikiEnabled,
				"snippets_enabled":                      project.SnippetsEnabled,
				"resolve_outdated_diff_discussions":     project.ResolveOutdatedDiffDiscussions,
				"container_registry_enabled":            project.ContainerRegistryEnabled,
				"created_at":                            project.CreatedAt.String(),
				"last_activity_at":                      project.LastActivityAt.String(),
				"creator_id":                            project.CreatorID,
				"namespace":                             flattenProjectNamespace(project.Namespace),
				"import_status":                         project.ImportStatus,
				"import_error":                          project.ImportError,
				"permissions":                           flattenProjectPermissions(project.Permissions),
				"archived":                              project.Archived,
				"avatar_url":                            project.AvatarURL,
				"shared_runners_enabled":                project.SharedRunnersEnabled,
				"forks_count":                           project.ForksCount,
				"star_count":                            project.StarCount,
				"runners_token":                         project.RunnersToken,
				"public_builds":                         project.PublicBuilds,
				"only_allow_merge_if_pipeline_succeeds": project.OnlyAllowMergeIfPipelineSucceeds,
				"only_allow_merge_if_all_discussions_are_resolved": project.OnlyAllowMergeIfAllDiscussionsAreResolved,
				"allow_merge_on_skipped_pipeline":                  project.AllowMergeOnSkippedPipeline,
				"lfs_enabled":                                      project.LFSEnabled,
				"request_access_enabled":                           project.RequestAccessEnabled,
				"merge_method":                                     project.MergeMethod,
				"forked_from_project":                              flattenForkedFromProject(project.ForkedFromProject),
				"mirror":                                           project.Mirror,
				"mirror_user_id":                                   project.MirrorUserID,
				"mirror_trigger_builds":                            project.MirrorTriggerBuilds,
				"only_mirror_protected_branches":                   project.OnlyMirrorProtectedBranches,
				"mirror_overwrites_diverged_branches":              project.MirrorOverwritesDivergedBranches,
				"shared_with_groups":                               flattenSharedWithGroupsOptions(project),
				"statistics":                                       project.Statistics,
				"_links":                                           flattenProjectLinks(project.Links),
				"ci_config_path":                                   project.CIConfigPath,
				"custom_attributes":                                project.CustomAttributes,
				"packages_enabled":                                 project.PackagesEnabled,
				"build_coverage_regex":                             project.BuildCoverageRegex,
			}
			values = append(values, v)
		}
	}
	return values
}

func dataSourceGitlabProjects() *schema.Resource {
	// lintignore: S024 // TODO: Resolve this tfproviderlint issue
	return &schema.Resource{
		Read: dataSourceGitlabProjectsRead,

		// lintignore: S006 // TODO: Resolve this tfproviderlint issue
		Schema: map[string]*schema.Schema{
			"max_queryable_pages": {
				Type:        schema.TypeInt,
				Description: "Prevents overloading your Gitlab instance in case of a misconfiguration.",
				Optional:    true,
				Default:     10,
			},
			"group_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"page": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"per_page": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntAtMost(100),
			},
			"archived": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"order_by": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"id",
					"name",
					"username",
					"created_at",
					"updated_at"}, true),
			},
			"sort": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"desc", "asc"}, true),
			},
			"search": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"simple": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"owned": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"starred": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"public",
					"private",
					"internal"}, true),
			},
			"with_issues_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"with_merge_requests_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"with_custom_attributes": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"membership": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"min_access_level": {
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: validation.IntInSlice([]int{
					int(gitlab.GuestPermissions),
					int(gitlab.DeveloperPermissions),
					int(gitlab.MaintainerPermissions),
					int(gitlab.MasterPermissions),
				}),
				ConflictsWith: []string{
					"group_id",
				},
			},
			"with_programming_language": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"group_id",
				},
			},
			"statistics": {
				Type:     schema.TypeBool,
				Optional: true,
				ConflictsWith: []string{
					"group_id",
				},
			},
			"with_shared": {
				Type:     schema.TypeBool,
				Optional: true,
				ConflictsWith: []string{
					"statistics",
					"with_programming_language",
					"min_access_level",
				},
			},
			"include_subgroups": {
				Type:     schema.TypeBool,
				Optional: true,
				ConflictsWith: []string{
					"statistics",
					"with_programming_language",
					"min_access_level",
				},
			},
			"projects": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"default_branch": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"visibility": {
							Type:     schema.TypeString,
							Computed: true,
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
						"readme_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tag_list": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"owner": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"username": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"state": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"avatar_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"website_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name_with_namespace": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path_with_namespace": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"issues_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"open_issues_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"merge_requests_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"approvals_before_merge": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"jobs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"wiki_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"snippets_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"resolve_outdated_diff_discussions": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"container_registry_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_activity_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creator_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"namespace": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"path": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"kind": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"full_path": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"import_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"import_error": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"permissions": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"project_access": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
									},
									"group_access": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
									},
								},
							},
						},
						"archived": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"avatar_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"shared_runners_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"forks_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"star_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"runners_token": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"public_builds": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"only_allow_merge_if_pipeline_succeeds": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"only_allow_merge_if_all_discussions_are_resolved": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"allow_merge_on_skipped_pipeline": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"lfs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"request_access_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"merge_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"forked_from_project": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_url_to_repo": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name_with_namespace": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"path": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"path_with_namespace": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"web_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"mirror": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"mirror_user_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"mirror_trigger_builds": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"only_mirror_protected_branches": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"mirror_overwrites_diverged_branches": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"shared_with_groups": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"group_access_level": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"group_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"statistics": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"_links": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ci_config_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"custom_attributes": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeMap,
							},
						},
						"packages_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"build_coverage_regex": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// CRUD methods

func dataSourceGitlabProjectsRead(d *schema.ResourceData, meta interface{}) (err error) {
	client := meta.(*gitlab.Client)
	var projectList []*gitlab.Project

	// Permanent parameters

	page := d.Get("page").(int)
	perPage := d.Get("per_page").(int)
	maxQueryablePages := d.Get("max_queryable_pages").(int)

	// Conditional parameters
	// Only way I found to conditionally pass a search parameter to the List(Group/Project)Options
	// Json marshalling a complete List(Group/Project)Options JSON object had conversion issues with booleans
	var archivedPtr *bool
	var includeSubGroupsPtr *bool
	var membershipPtr *bool
	var minAccessLevelPtr *gitlab.AccessLevelValue
	var orderByPtr *string
	var ownedPtr *bool
	var searchPtr *string
	var simplePtr *bool
	var sortPtr *string
	var starredPtr *bool
	var statisticsPtr *bool
	var visibilityPtr *gitlab.VisibilityValue
	var withCustomAttributesPtr *bool
	var withIssuesEnabledPtr *bool
	var withMergeRequestsEnabledPtr *bool
	var withProgrammingLanguagePtr *string
	var withSharedPtr *bool

	if data, ok := d.GetOk("archived"); ok {
		d := data.(bool)
		archivedPtr = &d
	}
	if data, ok := d.GetOk("include_subgroups"); ok {
		d := data.(bool)
		includeSubGroupsPtr = &d
	}
	if data, ok := d.GetOk("membership"); ok {
		d := data.(bool)
		membershipPtr = &d
	}
	if data, ok := d.GetOk("min_access_level"); ok {
		minAccessLevelPtr = gitlab.AccessLevel(gitlab.AccessLevelValue(data.(int)))
	}
	if data, ok := d.GetOk("order_by"); ok {
		d := data.(string)
		orderByPtr = &d
	}
	if data, ok := d.GetOk("owned"); ok {
		d := data.(bool)
		ownedPtr = &d
	}
	if data, ok := d.GetOk("search"); ok {
		d := data.(string)
		searchPtr = &d
	}
	if data, ok := d.GetOk("simple"); ok {
		d := data.(bool)
		simplePtr = &d
	}
	if data, ok := d.GetOk("sort"); ok {
		d := data.(string)
		sortPtr = &d
	}
	if data, ok := d.GetOk("starred"); ok {
		d := data.(bool)
		starredPtr = &d
	}
	if data, ok := d.GetOk("statistics"); ok {
		d := data.(bool)
		statisticsPtr = &d
	}
	if data, ok := d.GetOk("visibility"); ok {
		visibilityPtr = gitlab.Visibility(gitlab.VisibilityValue(data.(string)))
	}
	if data, ok := d.GetOk("with_custom_attributes"); ok {
		d := data.(bool)
		withCustomAttributesPtr = &d
	}
	if data, ok := d.GetOk("with_issues_enabled"); ok {
		d := data.(bool)
		withIssuesEnabledPtr = &d
	}
	if data, ok := d.GetOk("with_merge_requests_enabled"); ok {
		d := data.(bool)
		withMergeRequestsEnabledPtr = &d
	}
	if data, ok := d.GetOk("with_programming_language"); ok {
		d := data.(string)
		withProgrammingLanguagePtr = &d
	}
	if data, ok := d.GetOk("with_shared"); ok {
		d := data.(bool)
		withSharedPtr = &d
	}

	log.Printf("[DEBUG] Reading Gitlab projects")

	switch groupId, ok := d.GetOk("group_id"); ok {
	// GroupProject case
	case true:
		opts := &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
			Archived:                 archivedPtr,
			Visibility:               visibilityPtr,
			OrderBy:                  orderByPtr,
			Sort:                     sortPtr,
			Search:                   searchPtr,
			Simple:                   simplePtr,
			Owned:                    ownedPtr,
			Starred:                  starredPtr,
			WithIssuesEnabled:        withIssuesEnabledPtr,
			WithMergeRequestsEnabled: withMergeRequestsEnabledPtr,
			WithShared:               withSharedPtr,
			IncludeSubgroups:         includeSubGroupsPtr,
			WithCustomAttributes:     withCustomAttributesPtr,
		}

		for {
			projects, response, err := client.Groups.ListGroupProjects(groupId.(int), opts, nil)
			if err != nil {
				return err
			}
			projectList = append(projectList, projects...)
			opts.ListOptions.Page++

			log.Printf("[INFO] Currentpage: %d, Total: %d", response.CurrentPage, response.TotalPages)
			if response.CurrentPage == response.TotalPages || response.CurrentPage > maxQueryablePages {
				break
			}
		}
		h, err := hashstructure.Hash(*opts, nil)
		if err != nil {
			return err
		}
		d.SetId(fmt.Sprintf("%d-%d", groupId.(int), h))
		if err := d.Set("projects", flattenProjects(projectList)); err != nil {
			return err
		}

	// Project case
	default:
		opts := &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
			Archived:                 archivedPtr,
			OrderBy:                  orderByPtr,
			Sort:                     sortPtr,
			Search:                   searchPtr,
			Simple:                   simplePtr,
			Owned:                    ownedPtr,
			Membership:               membershipPtr,
			Starred:                  starredPtr,
			Statistics:               statisticsPtr,
			Visibility:               visibilityPtr,
			WithIssuesEnabled:        withIssuesEnabledPtr,
			WithMergeRequestsEnabled: withMergeRequestsEnabledPtr,
			MinAccessLevel:           minAccessLevelPtr,
			WithCustomAttributes:     withCustomAttributesPtr,
			WithProgrammingLanguage:  withProgrammingLanguagePtr,
		}

		for {
			projects, response, err := client.Projects.ListProjects(opts, nil)
			if err != nil {
				return err
			}
			projectList = append(projectList, projects...)
			opts.ListOptions.Page++

			log.Printf("[INFO] Currentpage: %d, Total: %d", response.CurrentPage, response.TotalPages)
			if response.CurrentPage == response.TotalPages || response.CurrentPage > maxQueryablePages {
				break
			}
		}
		h, err := hashstructure.Hash(*opts, nil)
		if err != nil {
			return err
		}
		d.SetId(fmt.Sprintf("%d", h))
		if err := d.Set("projects", flattenProjects(projectList)); err != nil {
			return err
		}
	}
	return err
}

func flattenSharedWithGroupsOptions(project *gitlab.Project) []interface{} {
	var sharedWithGroupsList []interface{}

	for _, option := range project.SharedWithGroups {
		values := map[string]interface{}{
			"group_id":           option.GroupID,
			"group_access_level": accessLevelValueToName[gitlab.AccessLevelValue(option.GroupAccessLevel)],
			"group_name":         option.GroupName,
		}

		sharedWithGroupsList = append(sharedWithGroupsList, values)
	}

	return sharedWithGroupsList
}
