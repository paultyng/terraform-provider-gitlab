package gitlab

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/xanzy/go-gitlab"
)

func resourceGitlabDeployToken() *schema.Resource {
	// lintignore: XR002 // TODO: Resolve this tfproviderlint issue
	return &schema.Resource{
		Create: resourceGitlabDeployTokenCreate,
		Read:   resourceGitlabDeployTokenRead,
		Delete: resourceGitlabDeployTokenDelete,

		Schema: map[string]*schema.Schema{
			"project": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"project", "group"},
				ForceNew:     true,
			},
			"group": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"project", "group"},
				ForceNew:     true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"expires_at": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.IsRFC3339Time,
				DiffSuppressFunc: expiresAtSuppressFunc,
				ForceNew:         true,
			},
			"scopes": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(
						[]string{
							"read_registry",
							"read_repository",
							"read_package_registry",
							"write_repository",
							"write_registry",
							"write_package_registry"
						}, false),
				},
			},

			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func expiresAtSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	oldDate, oldDateErr := time.Parse(time.RFC3339, old)
	newDate, newDateErr := time.Parse(time.RFC3339, new)
	if oldDateErr != nil || newDateErr != nil {
		return false
	}
	return oldDate == newDate
}

func resourceGitlabDeployTokenCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project, isProject := d.GetOk("project")
	group, isGroup := d.GetOk("group")

	var expiresAt *time.Time
	var err error

	if exp, ok := d.GetOk("expires_at"); ok {
		parsedExpiresAt, err := time.Parse(time.RFC3339, exp.(string))
		expiresAt = &parsedExpiresAt
		if err != nil {
			return fmt.Errorf("Invalid expires_at date: %v", err)
		}
	}

	scopes := stringSetToStringSlice(d.Get("scopes").(*schema.Set))

	var deployToken *gitlab.DeployToken

	if isProject {
		options := &gitlab.CreateProjectDeployTokenOptions{
			Name:      gitlab.String(d.Get("name").(string)),
			Username:  gitlab.String(d.Get("username").(string)),
			ExpiresAt: expiresAt,
			Scopes:    *scopes,
		}

		log.Printf("[DEBUG] Create GitLab deploy token %s in project %s", *options.Name, project.(string))

		deployToken, _, err = client.DeployTokens.CreateProjectDeployToken(project, options)

	} else if isGroup {
		options := &gitlab.CreateGroupDeployTokenOptions{
			Name:      gitlab.String(d.Get("name").(string)),
			Username:  gitlab.String(d.Get("username").(string)),
			ExpiresAt: expiresAt,
			Scopes:    *scopes,
		}

		log.Printf("[DEBUG] Create GitLab deploy token %s in group %s", *options.Name, group.(string))

		deployToken, _, err = client.DeployTokens.CreateGroupDeployToken(group, options)
	}

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", deployToken.ID))

	// Token is only available on creation
	d.Set("token", deployToken.Token)
	d.Set("username", deployToken.Username)

	return nil
}

func resourceGitlabDeployTokenRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project, isProject := d.GetOk("project")
	group, isGroup := d.GetOk("group")
	deployTokenID, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	var deployTokens []*gitlab.DeployToken

	if isProject {
		log.Printf("[DEBUG] Read GitLab deploy token %d in project %s", deployTokenID, project.(string))
		deployTokens, _, err = client.DeployTokens.ListProjectDeployTokens(project, nil)

	} else if isGroup {
		log.Printf("[DEBUG] Read GitLab deploy token %d in group %s", deployTokenID, group.(string))
		deployTokens, _, err = client.DeployTokens.ListGroupDeployTokens(group, nil)
	}
	if err != nil {
		return err
	}

	for _, token := range deployTokens {
		if token.ID == deployTokenID {
			d.Set("name", token.Name)
			d.Set("username", token.Username)

			if token.ExpiresAt != nil {
				d.Set("expires_at", token.ExpiresAt) // lintignore: R004,XR004 // TODO: Resolve this tfproviderlint issue
			}

			for _, scope := range token.Scopes {
				if scope == "read_repository" {
					d.Set("scopes.read_repository", true)
				}

				if scope == "read_registry" {
					d.Set("scopes.read_registry", true)
				}

				if scope == "read_package_registry" {
					d.Set("scopes.read_package_registry", true)
				}
                                if scope == "write_repository" {
                                        d.Set("scopes.write_repository", true)
                                }
                                if scope == "write_registry" {
                                        d.Set("scopes.write_registry", true)
                                }
                                if scope == "write_package_registry" {
                                        d.Set("scopes.write_package_registry", true)
                                }
			}
		}
	}

	return nil
}

func resourceGitlabDeployTokenDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project, isProject := d.GetOk("project")
	group, isGroup := d.GetOk("group")
	deployTokenID, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	var response *gitlab.Response

	if isProject {
		log.Printf("[DEBUG] Delete GitLab deploy token %d in project %s", deployTokenID, project.(string))
		response, err = client.DeployTokens.DeleteProjectDeployToken(project, deployTokenID)

	} else if isGroup {
		log.Printf("[DEBUG] Delete GitLab deploy token %d in group %s", deployTokenID, group.(string))
		response, err = client.DeployTokens.DeleteGroupDeployToken(group, deployTokenID)
	}
	if err != nil {
		return err
	}

	// StatusNoContent = 204
	// Success with no body
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Invalid status code returned: %s", response.Status)
	}

	return nil
}
