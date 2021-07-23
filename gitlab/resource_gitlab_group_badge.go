package gitlab

import (
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabGroupBadge() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabGroupBadgeCreate,
		Read:   resourceGitlabGroupBadgeRead,
		Update: resourceGitlabGroupBadgeUpdate,
		Delete: resourceGitlabGroupBadgeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"group": {
				Type:     schema.TypeString,
				Required: true,
			},
			"link_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rendered_link_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rendered_image_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGitlabGroupBadgeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	groupID := d.Get("group").(string)
	options := &gitlab.AddGroupBadgeOptions{
		LinkURL:  gitlab.String(d.Get("link_url").(string)),
		ImageURL: gitlab.String(d.Get("image_url").(string)),
	}

	log.Printf("[DEBUG] create gitlab group variable %s/%s", *options.LinkURL, *options.ImageURL)

	badge, _, err := client.GroupBadges.AddGroupBadge(groupID, options)
	if err != nil {
		return err
	}

	badgeID := strconv.Itoa(badge.ID)

	d.SetId(buildTwoPartID(&groupID, &badgeID))

	return resourceGitlabGroupBadgeRead(d, meta)
}

func resourceGitlabGroupBadgeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	ids := strings.Split(d.Id(), ":")
	groupID := ids[0]
	badgeID, err := strconv.Atoi(ids[1])
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] read gitlab group badge %s/%d", groupID, badgeID)

	badge, _, err := client.GroupBadges.GetGroupBadge(groupID, badgeID)
	if err != nil {
		return err
	}

	resourceGitlabGroupBadgeSetToState(d, badge, &groupID)
	return nil
}

func resourceGitlabGroupBadgeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	ids := strings.Split(d.Id(), ":")
	groupID := ids[0]
	badgeID, err := strconv.Atoi(ids[1])
	if err != nil {
		return err
	}

	options := &gitlab.EditGroupBadgeOptions{
		LinkURL:  gitlab.String(d.Get("link_url").(string)),
		ImageURL: gitlab.String(d.Get("image_url").(string)),
	}

	log.Printf("[DEBUG] update gitlab group badge %s/%d", groupID, badgeID)

	_, _, err = client.GroupBadges.EditGroupBadge(groupID, badgeID, options)
	if err != nil {
		return err
	}

	return resourceGitlabGroupBadgeRead(d, meta)
}

func resourceGitlabGroupBadgeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	ids := strings.Split(d.Id(), ":")
	groupID := ids[0]
	badgeID, err := strconv.Atoi(ids[1])
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Delete gitlab group badge %s/%d", groupID, badgeID)

	_, err = client.GroupBadges.DeleteGroupBadge(groupID, badgeID)
	return err
}

func resourceGitlabGroupBadgeSetToState(d *schema.ResourceData, badge *gitlab.GroupBadge, groupID *string) {
	d.Set("link_url", badge.LinkURL)
	d.Set("image_url", badge.ImageURL)
	d.Set("rendered_link_url", badge.RenderedLinkURL)
	d.Set("rendered_image_url", badge.RenderedImageURL)
	d.Set("group", groupID)
}
