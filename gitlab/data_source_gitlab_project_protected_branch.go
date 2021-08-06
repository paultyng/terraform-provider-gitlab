package gitlab

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/xanzy/go-gitlab"
)

func dataSourceGitlabProjectProtectedBranch() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGitlabProjectProtectedBranchRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeString,
				Description:  "ID or URL encoded name of project",
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "Name of the protected branch",
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
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
	}
}

func dataSourceGitlabProjectProtectedBranchSchemaAccessLevels() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"access_level": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"access_level_description": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"user_id": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"group_id": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			},
		},
	}
}

func dataSourceGitlabProjectProtectedBranchRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	log.Printf("[INFO] Reading Gitlab protected branch")

	project := d.Get("project_id")
	name := d.Get("name").(string)

	// Get protected branch by project ID/path and branch name
	pb, _, err := client.ProtectedBranches.GetProtectedBranch(project, name)
	if err != nil {
		return fmt.Errorf("error getting protected branch (Project: %v / Name %v): %v", project, name, err)
	}

	if err := d.Set("push_access_levels", convertBranchAccessDescriptionsToStateBranchAccessDescriptions(pb.PushAccessLevels)); err != nil {
		return err
	}
	if err := d.Set("merge_access_levels", convertBranchAccessDescriptionsToStateBranchAccessDescriptions(pb.MergeAccessLevels)); err != nil {
		return err
	}
	if err := d.Set("unprotect_access_levels", convertBranchAccessDescriptionsToStateBranchAccessDescriptions(pb.UnprotectAccessLevels)); err != nil {
		return err
	}
	if err := d.Set("code_owner_approval_required", pb.CodeOwnerApprovalRequired); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", pb.ID))

	return nil
}

type stateBranchAccessDescription struct {
	AccessLevel            string `json:"access_level" mapstructure:"access_level"`
	AccessLevelDescription string `json:"access_level_description" mapstructure:"access_level_description"`
	GroupID                *int   `json:"group_id,omitempty" mapstructure:"group_id,omitempty"`
	UserID                 *int   `json:"user_id,omitempty" mapstructure:"user_id,omitempty"`
}

func convertBranchAccessDescriptionsToStateBranchAccessDescriptions(descriptions []*gitlab.BranchAccessDescription) []stateBranchAccessDescription {
	result := make([]stateBranchAccessDescription, 0)

	for _, description := range descriptions {
		result = append(result, convertBranchAccessDescriptionToStateBranchAccessDescription(description))
	}

	return result
}

func convertBranchAccessDescriptionToStateBranchAccessDescription(description *gitlab.BranchAccessDescription) stateBranchAccessDescription {
	stateDescription := stateBranchAccessDescription{
		AccessLevel:            accessLevel[description.AccessLevel],
		AccessLevelDescription: description.AccessLevelDescription,
	}
	if description.UserID != 0 {
		stateDescription.UserID = &description.UserID
	}
	if description.GroupID != 0 {
		stateDescription.GroupID = &description.GroupID
	}
	return stateDescription
}
