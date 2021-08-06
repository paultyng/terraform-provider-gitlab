package gitlab

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabUserImpersonationToken() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabUserImpersonationTokenCreate,
		Read:   resourceGitlabUserImpersonationTokenRead,
		Delete: resourceGitlabUserImpersonationTokenDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"revoked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"token": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"scopes": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"api", "read_user"}, true),
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires_at": {
				Type:         schema.TypeString,
				ValidateFunc: validateDateFunc,
				Optional:     true,
				ForceNew:     true,
			},
		},
	}
}

func resourceGitlabUserImpersonationTokenCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	user_id := d.Get("user_id").(int)
	options := &gitlab.CreateImpersonationTokenOptions{
		Name: gitlab.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("scopes"); ok {
		options.Scopes = stringSetToStringSlice(v.(*schema.Set))
	}

	if v, ok := d.GetOk("expires_at"); ok {
		layout := "2006-01-02"
		t, _ := time.Parse(layout, v.(string))
		options.ExpiresAt = &t
	}

	log.Printf("[DEBUG] create gitlab user impersonation %s for user %d", *options.Name, user_id)
	impersonationToken, _, err := client.Users.CreateImpersonationToken(user_id, options)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d:%d", user_id, impersonationToken.ID))
	// We need to set token here instead of read, as it is only returned once
	d.Set("token", impersonationToken.Token)

	return resourceGitlabUserImpersonationTokenRead(d, meta)
}

func resourceGitlabUserImpersonationTokenRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	usertoken := strings.Split(d.Id(), ":")
	userId, _ := strconv.Atoi(usertoken[0])
	tokenId, _ := strconv.Atoi(usertoken[1])

	log.Printf("[DEBUG] read gitlab user impersonation token %d:%d", userId, tokenId)

	impersonationToken, _, err := client.Users.GetImpersonationToken(userId, tokenId)
	if err != nil {
		return err
	}

	d.Set("user_id", userId)
	d.Set("name", impersonationToken.Name)
	d.Set("active", impersonationToken.Active)
	d.Set("revoked", impersonationToken.Revoked)
	d.Set("scopes", impersonationToken.Scopes)
	d.Set("created_at", impersonationToken.CreatedAt)
	d.Set("expires_at", impersonationToken.ExpiresAt)
	return nil
}

// In case we delete an impersonation token, Gitlab will update it with `revoked: true`
// so the object still exists on Gitlab side, but we remove it from TF state
func resourceGitlabUserImpersonationTokenDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	usertoken := strings.Split(d.Id(), ":")
	userId, _ := strconv.Atoi(usertoken[0])
	tokenId, _ := strconv.Atoi(usertoken[1])

	log.Printf("[DEBUG] delete (revoke) gitlab user impersonation token %d:%d", userId, tokenId)

	_, err := client.Users.RevokeImpersonationToken(userId, tokenId)
	return err
}
