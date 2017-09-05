package gitlab

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabUserCreate,
		Read:   resourceGitlabUserRead,
		Update: resourceGitlabUserUpdate,
		Delete: resourceGitlabUserDelete,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_admin": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"can_create_group": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"confirm": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"projects_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
		},
	}
}

func resourceGitlabUserSetToState(d *schema.ResourceData, user *gitlab.User) {
	d.Set("username", user.Username)
	d.Set("name", user.Name)
	d.Set("can_create_group", user.CanCreateGroup)
	d.Set("projects_limit", user.ProjectsLimit)
}

func resourceGitlabUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	options := &gitlab.CreateUserOptions{
		Email:          gitlab.String(d.Get("email").(string)),
		Password:       gitlab.String(d.Get("password").(string)),
		Username:       gitlab.String(d.Get("username").(string)),
		Name:           gitlab.String(d.Get("name").(string)),
		ProjectsLimit:  gitlab.Int(d.Get("projects_limit").(int)),
		Admin:          gitlab.Bool(d.Get("is_admin").(bool)),
		CanCreateGroup: gitlab.Bool(d.Get("can_create_group").(bool)),
		Confirm:        gitlab.Bool(d.Get("confirm").(bool)),
	}

	log.Printf("[DEBUG] create gitlab user %q", options.Username)

	user, _, err := client.Users.CreateUser(options)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", user.ID))
	d.Set("is_admin", user.IsAdmin)

	return resourceGitlabUserRead(d, meta)
}

func resourceGitlabUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] read gitlab user %s", d.Id())

	id, _ := strconv.Atoi(d.Id())

	user, response, err := client.Users.GetUser(id)
	if err != nil {
		if response.StatusCode == 404 {
			log.Printf("[WARN] removing user %s from state because it no longer exists in gitlab", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	resourceGitlabUserSetToState(d, user)
	return nil
}

func resourceGitlabUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	options := &gitlab.ModifyUserOptions{}

	if d.HasChange("name") {
		options.Name = gitlab.String(d.Get("name").(string))
	}

	if d.HasChange("username") {
		options.Username = gitlab.String(d.Get("username").(string))
	}

	if d.HasChange("is_admin") {
		options.Admin = gitlab.Bool(d.Get("is_admin").(bool))
	}

	if d.HasChange("can_create_group") {
		options.CanCreateGroup = gitlab.Bool(d.Get("can_create_group").(bool))
	}

	if d.HasChange("projects_limit") {
		options.ProjectsLimit = gitlab.Int(d.Get("projects_limit").(int))
	}

	log.Printf("[DEBUG] update gitlab user %s", d.Id())

	id, _ := strconv.Atoi(d.Id())

	_, _, err := client.Users.ModifyUser(id, options)
	if err != nil {
		return err
	}

	return resourceGitlabUserRead(d, meta)
}

func resourceGitlabUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] Delete gitlab user %s", d.Id())

	id, _ := strconv.Atoi(d.Id())

	_, err := client.Users.DeleteUser(id)
	// Ignoring error due to some bug in library
	log.Printf("[DEBUG] Delete gitlab user %s", err)
	return nil
}
