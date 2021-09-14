package gitlab

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabUserSSHKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabUserSSHKeyCreate,
		Read:   resourceGitlabUserSSHKeyRead,
		Update: resourceGitlabUserSSHKeyUpdate,
		Delete: resourceGitlabUserSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGitlabUserSSHKeyImporter,
		},

		Schema: map[string]*schema.Schema{
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceGitlabUserSSHKeyImporter(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	s := strings.Split(d.Id(), ":")
	if len(s) != 2 {
		d.SetId("")
		return nil, fmt.Errorf("Invalid SSH Key import format; expected '{user_id}:{key_id}'")
	}

	userID, err := strconv.Atoi(s[0])
	if err != nil {
		return nil, fmt.Errorf("Invalid SSH Key import format; expected '{user_id}:{key_id}'")
	}

	d.Set("user_id", userID)
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}

func resourceGitlabUserSSHKeySetToState(d *schema.ResourceData, key *gitlab.SSHKey) {
	d.Set("title", key.Title)
	d.Set("key", key.Key)
	d.Set("created_at", key.CreatedAt)
}

func resourceGitlabUserSSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	userID := d.Get("user_id").(int)

	options := &gitlab.AddSSHKeyOptions{
		Title: gitlab.String(d.Get("title").(string)),
		Key:   gitlab.String(d.Get("key").(string)),
	}

	key, _, err := client.Users.AddSSHKeyForUser(userID, options)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", key.ID))

	return resourceGitlabUserSSHKeyRead(d, meta)
}

func resourceGitlabUserSSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	id, _ := strconv.Atoi(d.Id())
	userID := d.Get("user_id").(int)

	keys, _, err := client.Users.ListSSHKeysForUser(userID, &gitlab.ListSSHKeysForUserOptions{})
	if err != nil {
		return err
	}

	var key *gitlab.SSHKey

	for _, k := range keys {
		if k.ID == id {
			key = k
			break
		}
	}

	if key == nil {
		return fmt.Errorf("Could not find sshkey %d for user %d", id, userID)
	}

	resourceGitlabUserSSHKeySetToState(d, key)
	return nil
}

func resourceGitlabUserSSHKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := resourceGitlabUserSSHKeyDelete(d, meta); err != nil {
		return err
	}
	if err := resourceGitlabUserSSHKeyCreate(d, meta); err != nil {
		return err
	}
	return resourceGitlabUserSSHKeyRead(d, meta)
}

func resourceGitlabUserSSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] Delete gitlab user sshkey %s", d.Id())

	id, _ := strconv.Atoi(d.Id())
	userID := d.Get("user_id").(int)

	if _, err := client.Users.DeleteSSHKeyForUser(userID, id); err != nil {
		return err
	}

	return nil
}
