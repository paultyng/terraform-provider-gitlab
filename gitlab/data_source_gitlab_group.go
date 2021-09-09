package gitlab

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/xanzy/go-gitlab"
)

func dataSourceGitlabGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGitlabGroupRead,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
				ConflictsWith: []string{
					"full_path",
				},
			},
			"full_path": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ConflictsWith: []string{
					"group_id",
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"web_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
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
			"visibility_level": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"runners_token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"default_branch_protection": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceGitlabGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	var group *gitlab.Group
	var err error

	log.Printf("[INFO] Reading Gitlab group")

	groupIDData, groupIDOk := d.GetOk("group_id")
	fullPathData, fullPathOk := d.GetOk("full_path")

	if groupIDOk {
		// Get group by id
		group, _, err = client.Groups.GetGroup(groupIDData.(int), nil)
		if err != nil {
			return err
		}
	} else if fullPathOk {
		// Get group by full path
		group, _, err = client.Groups.GetGroup(fullPathData.(string), nil)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("one and only one of group_id or full_path must be set")
	}

	d.Set("group_id", group.ID)
	d.Set("full_path", group.FullPath)
	d.Set("name", group.Name)
	d.Set("full_name", group.FullName)
	d.Set("web_url", group.WebURL)
	d.Set("path", group.Path)
	d.Set("description", group.Description)
	d.Set("lfs_enabled", group.LFSEnabled)
	d.Set("request_access_enabled", group.RequestAccessEnabled)
	d.Set("visibility_level", group.Visibility)
	d.Set("parent_id", group.ParentID)
	d.Set("runners_token", group.RunnersToken)
	d.Set("default_branch_protection", group.DefaultBranchProtection)

	d.SetId(fmt.Sprintf("%d", group.ID))

	return nil
}
