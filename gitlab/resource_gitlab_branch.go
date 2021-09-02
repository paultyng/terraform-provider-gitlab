package gitlab

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
	"log"
)

func resourceGitlabBranch() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabBranchCreate,
		Read:   resourceGitlabBranchRead,
		Delete: resourceGitlabBranchDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"project": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"ref": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "main", // Default value required for import logic -- api does not return consistent value to use for ref
			},
			"web_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"can_push": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"merged": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"commit": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      schema.HashResource(commitSchema),
				Elem:     commitSchema,
			},
		},
	}
}

var commitSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"author_email": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"author_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"authored_date": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"committed_date": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"committer_email": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"committer_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"short_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"title": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"message": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"parent_ids": {
			Type:     schema.TypeSet,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Set:      schema.HashString,
		},
	},
}

func resourceGitlabBranchCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	name := d.Get("name").(string)
	project := d.Get("project").(string)
	ref := d.Get("ref").(string)
	branchOptions := &gitlab.CreateBranchOptions{
		Branch: &name, Ref: &ref,
	}

	log.Printf("[DEBUG] create gitlab branch %s for project %s with ref %s", name, project, ref)
	branch, resp, err := client.Branches.CreateBranch(project, branchOptions)
	if err != nil {
		log.Printf("[DEBUG] failed to create gitlab branch %v response %v", branch, resp)
		return err
	}
	d.SetId(buildTwoPartID(&project, &name))
	return resourceGitlabBranchRead(d, meta)
}

func resourceGitlabBranchRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project, name, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] read gitlab branch %s", name)
	branch, resp, err := client.Branches.GetBranch(project, name)
	if err != nil {
		if resp.StatusCode == 404 {
			log.Printf("[DEBUG] recieved 404 for gitlab branch %s, removing from state", name)
			d.SetId("")
			return err
		}
		log.Printf("[DEBUG] failed to read gitlab branch %s response %v", name, resp)
		return err
	}
	ref := d.Get("ref").(string)
	if ref == "" {
		ref = "main" // on import this will be set
	}
	d.SetId(buildTwoPartID(&project, &name))
	d.Set("name", branch.Name)
	d.Set("project", project)
	d.Set("ref", ref)
	d.Set("web_url", branch.WebURL)
	d.Set("default", branch.Default)
	d.Set("can_push", branch.CanPush)
	d.Set("merged", branch.Merged)
	commit := flattenCommit(branch.Commit)
	if err := d.Set("commit", commit); err != nil {
		return err
	}

	return nil
}

func resourceGitlabBranchDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	name := d.Get("name").(string)
	log.Printf("[DEBUG] delete gitlab branch %s", name)
	resp, err := client.Branches.DeleteBranch(project, name)
	if err != nil {
		log.Printf("[DEBUG] failed to delete gitlab branch %s response %v", name, resp)
	}
	return err
}

func flattenCommit(commit *gitlab.Commit) (values []map[string]interface{}) {
	if commit == nil {
		return []map[string]interface{}{}
	}
	return []map[string]interface{}{
		{
			"id":              commit.ID,
			"short_id":        commit.ShortID,
			"title":           commit.Title,
			"author_name":     commit.AuthorName,
			"author_email":    commit.AuthorEmail,
			"authored_date":   commit.AuthoredDate.String(),
			"committed_date":  commit.CommittedDate.String(),
			"committer_email": commit.CommitterEmail,
			"committer_name":  commit.CommitterName,
			"message":         commit.Message,
			"parent_ids":      commit.ParentIDs,
		},
	}
}
