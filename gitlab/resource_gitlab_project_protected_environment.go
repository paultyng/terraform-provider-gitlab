package gitlab

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	gitlab "github.com/xanzy/go-gitlab"
)

// https://docs.gitlab.com/ee/ci/environments/protected_environments.html
func resourceGitlabProjectProtectedEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabProjectProtectedEnvironmentCreate,
		Read:   resourceGitlabProjectProtectedEnvironmentRead,
		Delete: resourceGitlabProjectProtectedEnvironmentDelete,
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
			"environment": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"deploy_access_levels": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_level": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"developer", "maintainer"}, false),
						},
						"access_level_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_id": {
							Type:         schema.TypeInt,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"group_id": {
							Type:         schema.TypeInt,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
		},
	}
}

func resourceGitlabProjectProtectedEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	deployAccessLevels, err := expandDeployAccessLevels(d.Get("deploy_access_levels").([]interface{}))
	if err != nil {
		return fmt.Errorf("error expanding deploy_access_levels: %v", err)
	}
	environment := d.Get("environment").(string)
	options := gitlab.ProtectRepositoryEnvironmentsOptions{
		Name:               &environment,
		DeployAccessLevels: deployAccessLevels,
	}

	project := d.Get("project").(string)

	log.Printf("[DEBUG] Project %s create gitlab protected environment %q", project, *options.Name)

	client := meta.(*gitlab.Client)

	protectedEnvironment, resp, err := client.ProtectedEnvironments.ProtectRepositoryEnvironments(project, &options)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("feature Protected Environments is not available")
		}
		return err
	}

	d.SetId(buildTwoPartID(&project, &protectedEnvironment.Name))

	return resourceGitlabProjectProtectedEnvironmentRead(d, meta)
}

func resourceGitlabProjectProtectedEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] read gitlab protected environment %s", d.Id())

	project, environment, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}
	d.Set("project", project)
	d.Set("environment", environment)

	log.Printf("[DEBUG] Project %s read gitlab protected environment %q", project, environment)

	client := meta.(*gitlab.Client)

	protectedEnvironment, resp, err := client.ProtectedEnvironments.GetProtectedEnvironment(project, environment)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] Project %s gitlab protected environment %q not found", project, environment)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error getting gitlab project %q protected environment %q: %v", project, environment, err)
	}

	d.Set("environment", protectedEnvironment.Name)
	if err := d.Set("deploy_access_levels", flattenDeployAccessLevels(protectedEnvironment.DeployAccessLevels)); err != nil {
		return fmt.Errorf("error setting deploy_access_levels: %v", err)
	}

	return nil
}

func resourceGitlabProjectProtectedEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	project, environmentName, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Project %s delete gitlab project-level protected environment %s", project, environmentName)

	client := meta.(*gitlab.Client)

	_, err = client.ProtectedEnvironments.UnprotectEnvironment(project, environmentName)
	if err != nil {
		return err
	}

	return nil
}

func expandDeployAccessLevels(vs []interface{}) ([]*gitlab.EnvironmentAccessOptions, error) {
	result := make([]*gitlab.EnvironmentAccessOptions, 0)

	for _, v := range vs {
		opts := v.(map[string]interface{})
		option := &gitlab.EnvironmentAccessOptions{}
		if accessLevel, exists := opts["access_level"]; exists {
			option.AccessLevel = gitlab.AccessLevel(accessLevelNameToValue[accessLevel.(string)])
		} else if userID, exists := opts["user_id"]; exists {
			option.UserID = gitlab.Int(userID.(int))
		} else if groupID, exists := opts["group_id"]; exists {
			option.GroupID = gitlab.Int(groupID.(int))
		}
		result = append(result, option)
	}

	return result, nil
}

func flattenDeployAccessLevels(vs []*gitlab.EnvironmentAccessDescription) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, accessDescription := range vs {
		v := make(map[string]interface{})
		v["access_level"] = accessLevelValueToName[accessDescription.AccessLevel]
		v["access_level_description"] = accessDescription.AccessLevelDescription
		if accessDescription.UserID != 0 {
			v["user_id"] = accessDescription.UserID
		}
		if accessDescription.GroupID != 0 {
			v["group_id"] = accessDescription.GroupID
		}
		result = append(result, v)
	}

	return result
}
