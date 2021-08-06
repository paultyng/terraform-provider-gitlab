package gitlab

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabDeployEnableKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabDeployKeyEnableCreate,
		Read:   resourceGitlabDeployKeyEnableRead,
		Delete: resourceGitlabDeployKeyEnableDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGitlabDeployKeyEnableStateImporter,
		},

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"title": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"can_push": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceGitlabDeployKeyEnableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	key_id, err := strconv.Atoi(d.Get("key_id").(string)) // nolint // TODO: Resolve this golangci-lint issue: ineffectual assignment to err (ineffassign)

	log.Printf("[DEBUG] enable gitlab deploy key %s/%d", project, key_id)

	deployKey, _, err := client.DeployKeys.EnableDeployKey(project, key_id)
	if err != nil {
		return err
	}

	options := &gitlab.UpdateDeployKeyOptions{
		CanPush: gitlab.Bool(d.Get("can_push").(bool)),
	}
	_, _, err = client.DeployKeys.UpdateDeployKey(project, key_id, options)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%d", project, deployKey.ID))

	return resourceGitlabDeployKeyEnableRead(d, meta)
}

func resourceGitlabDeployKeyEnableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	deployKeyID, err := strconv.Atoi(d.Get("key_id").(string))
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] read gitlab deploy key %s/%d", project, deployKeyID)

	deployKey, _, err := client.DeployKeys.GetDeployKey(project, deployKeyID)
	if err != nil {
		return err
	}

	d.Set("title", deployKey.Title)
	d.Set("key_id", deployKey.ID)
	d.Set("key", deployKey.Key)
	d.Set("can_push", deployKey.CanPush)
	d.Set("project", project)
	return nil
}

func resourceGitlabDeployKeyEnableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	deployKeyID, err := strconv.Atoi(d.Get("key_id").(string))
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Delete gitlab deploy key %s/%d", project, deployKeyID)

	response, err := client.DeployKeys.DeleteDeployKey(project, deployKeyID)

	// HTTP 2XX is success including 204 with no body
	if response.StatusCode/100 == 2 {
		return nil
	}
	return err
}

func resourceGitlabDeployKeyEnableStateImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := strings.Split(d.Id(), ":")
	if len(s) != 2 {
		d.SetId("")
		return nil, fmt.Errorf("Invalid Deploy Key import format; expected '{project_id}:{deploy_key_id}'")
	}
	project, id := s[0], s[1]

	d.SetId(fmt.Sprintf("%s:%s", project, id))
	d.Set("key_id", id)
	d.Set("project", project)

	return []*schema.ResourceData{d}, nil
}
