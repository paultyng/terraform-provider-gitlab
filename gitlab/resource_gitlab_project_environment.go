package gitlab

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	gitlab "github.com/xanzy/go-gitlab"
)

// https://docs.gitlab.com/ee/ci/environments/protected_environments.html
func resourceGitlabProjectEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabProjectEnvironmentCreate,
		Read:   resourceGitlabProjectEnvironmentRead,
		Update: resourceGitlabProjectEnvironmentUpdate,
		Delete: resourceGitlabProjectEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"project": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"external_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGitlabProjectEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	options := gitlab.CreateEnvironmentOptions{
		Name: &name,
	}
	if externalURL, ok := d.GetOk("external_url"); ok {
		options.ExternalURL = gitlab.String(externalURL.(string))
	}

	project := d.Get("project").(string)

	log.Printf("[DEBUG] Project %s create gitlab environment %q", project, *options.Name)

	client := meta.(*gitlab.Client)

	environment, resp, err := client.Environments.CreateEnvironment(project, &options)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("feature Environments is not available")
		}
		return err
	}

	d.SetId(buildTwoPartID(&project, gitlab.String(fmt.Sprintf("%v", environment.ID))))

	return resourceGitlabProjectEnvironmentRead(d, meta)
}

func resourceGitlabProjectEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] read gitlab environment %s", d.Id())

	project, environmentID, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}
	environmentIDInt, err := strconv.Atoi(environmentID)

	log.Printf("[DEBUG] Project %s read gitlab environment %d", project, environmentIDInt)

	client := meta.(*gitlab.Client)

	environment, resp, err := client.Environments.GetEnvironment(project, environmentIDInt)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] Project %s gitlab environment %q not found", project, environmentID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error getting gitlab project %q environment %q: %v", project, environmentID, err)
	}

	d.SetId(buildTwoPartID(&project, gitlab.String(fmt.Sprintf("%v", environment.ID))))
	d.Set("project", project)
	d.Set("name", environment.Name)
	d.Set("state", environment.State)

	return nil
}

func resourceGitlabProjectEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] update gitlab environment %s", d.Id())

	project, environmentID, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}
	environmentIDInt, err := strconv.Atoi(environmentID)
	if err != nil {
		return fmt.Errorf("error converting environment ID to int: %v", err)
	}

	name := d.Get("name").(string)
	options := gitlab.EditEnvironmentOptions{
		Name: &name,
	}
	if d.HasChange("external_url") {
		options.ExternalURL = gitlab.String(d.Get("external_url").(string))
	}

	log.Printf("[DEBUG] Project %s update gitlab environment %d", project, environmentIDInt)

	client := meta.(*gitlab.Client)

	if _, _, err := client.Environments.EditEnvironment(project, environmentIDInt, &options); err != nil {
		return fmt.Errorf("error editing gitlab project %q environment %q: %v", project, environmentID, err)
	}

	return resourceGitlabProjectEnvironmentRead(d, meta)
}

func resourceGitlabProjectEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	project, environmentID, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}
	environmentIDInt, err := strconv.Atoi(environmentID)
	if err != nil {
		return fmt.Errorf("error converting environment ID to int: %v", err)
	}

	log.Printf("[DEBUG] Project %s delete gitlab project-level environment %v", project, environmentIDInt)

	client := meta.(*gitlab.Client)

	_, err = client.Environments.StopEnvironment(project, environmentIDInt)
	if err != nil {
		return fmt.Errorf("error stopping gitlab project %q environment %q: %v", project, environmentID, err)
	}
	_, err = client.Environments.DeleteEnvironment(project, environmentIDInt)
	if err != nil {
		return fmt.Errorf("error deleting gitlab project %q environment %q: %v", project, environmentID, err)
	}

	return nil
}
