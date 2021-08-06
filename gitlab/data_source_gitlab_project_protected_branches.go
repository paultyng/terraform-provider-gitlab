package gitlab

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mitchellh/hashstructure"
	"github.com/xanzy/go-gitlab"
)

func dataSourceGitlabProjectProtectedBranches() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGitlabProjectProtectedBranchesRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Description: "ID or URL encoded name of project",
				Required:    true,
			},
			"protected_branches": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the protected branch",
							Computed:    true,
						},
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"push_access_levels":      dataSourceGitlabProjectProtectedBranchSchemaAccessLevels(),
						"merge_access_levels":     dataSourceGitlabProjectProtectedBranchSchemaAccessLevels(),
						"unprotect_access_levels": dataSourceGitlabProjectProtectedBranchSchemaAccessLevels(),
						"code_owner_approval_required": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

type stateProtectedBranch struct {
	ID                        int                            `json:"id,omitempty" mapstructure:"id,omitempty"`
	Name                      string                         `json:"name,omitempty" mapstructure:"name,omitempty"`
	PushAccessLevels          []stateBranchAccessDescription `json:"push_access_levels,omitempty" mapstructure:"push_access_levels,omitempty"`
	MergeAccessLevels         []stateBranchAccessDescription `json:"merge_access_levels,omitempty" mapstructure:"merge_access_levels,omitempty"`
	UnprotectAccessLevels     []stateBranchAccessDescription `json:"unprotect_access_levels,omitempty" mapstructure:"unprotect_access_levels,omitempty"`
	CodeOwnerApprovalRequired bool                           `json:"code_owner_approval_required,omitempty" mapstructure:"code_owner_approval_required,omitempty"`
}

func dataSourceGitlabProjectProtectedBranchesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	log.Printf("[INFO] Reading Gitlab protected branch")

	project := d.Get("project_id")

	projectObject, _, err := client.Projects.GetProject(project, &gitlab.GetProjectOptions{})
	if err != nil {
		return err
	}

	allProtectedBranches := make([]stateProtectedBranch, 0)
	totalPages := -1
	opts := &gitlab.ListProtectedBranchesOptions{}
	for opts.Page = 0; opts.Page != totalPages; opts.Page++ {
		// Get protected branch by project ID/path and branch name
		pbs, resp, err := client.ProtectedBranches.ListProtectedBranches(project, opts)
		if err != nil {
			return err
		}
		totalPages = resp.TotalPages
		for _, pb := range pbs {
			allProtectedBranches = append(allProtectedBranches, stateProtectedBranch{
				ID:                        pb.ID,
				Name:                      pb.Name,
				PushAccessLevels:          convertBranchAccessDescriptionsToStateBranchAccessDescriptions(pb.PushAccessLevels),
				MergeAccessLevels:         convertBranchAccessDescriptionsToStateBranchAccessDescriptions(pb.MergeAccessLevels),
				UnprotectAccessLevels:     convertBranchAccessDescriptionsToStateBranchAccessDescriptions(pb.UnprotectAccessLevels),
				CodeOwnerApprovalRequired: pb.CodeOwnerApprovalRequired,
			})
		}
	}

	if err := d.Set("protected_branches", allProtectedBranches); err != nil {
		return err
	}

	h, err := hashstructure.Hash(*opts, nil)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d-%d", projectObject.ID, h))

	return nil
}
