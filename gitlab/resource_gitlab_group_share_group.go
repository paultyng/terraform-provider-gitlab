package gitlab

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

// https://docs.gitlab.com/ee/api/groups.html#share-groups-with-groups

func resourceGitlabGroupShareGroup() *schema.Resource {
	acceptedAccessLevels := make([]string, 0, len(accessLevelID))
	for k := range accessLevelID {
		acceptedAccessLevels = append(acceptedAccessLevels, k)
	}

	return &schema.Resource{
		Create: resourceGitlabGroupShareGroupCreate,
		Read:   resourceGitlabGroupShareGroupRead,
		Delete: resourceGitlabGroupShareGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"share_group_id": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Required: true,
			},
			"group_access": {
				Type:         schema.TypeString,
				ValidateFunc: validateValueFunc(acceptedAccessLevels),
				ForceNew:     true,
				Required:     true,
			},
			"expires_at": {
				Type:         schema.TypeString, // Format YYYY-MM-DD
				ValidateFunc: validateDateFunc,
				ForceNew:     true,
				Optional:     true,
			},
		},
	}
}

func resourceGitlabGroupShareGroupCreate(d *schema.ResourceData, meta interface{}) error {
	groupId := d.Get("group_id").(string)
	shareGroupId := d.Get("share_group_id").(int)
	groupAccess := accessLevelID[d.Get("group_access").(string)]
	options := &gitlab.ShareWithGroupOptions{
		GroupID:     &shareGroupId,
		GroupAccess: &groupAccess,
		ExpiresAt:   gitlab.String(d.Get("expires_at").(string)),
	}

	client := meta.(*gitlab.Client)
	log.Printf("[DEBUG] create gitlab group share for %d in %s", shareGroupId, groupId)

	_, _, err := client.GroupMembers.ShareWithGroup(groupId, options)
	if err != nil {
		return err
	}

	shareGroupIdString := strconv.Itoa(shareGroupId)
	d.SetId(buildTwoPartID(&groupId, &shareGroupIdString))

	return resourceGitlabGroupShareGroupRead(d, meta)
}

func resourceGitlabGroupShareGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	id := d.Id()
	log.Printf("[DEBUG] read gitlab shared groups %s", id)

	groupId, sharedGroupId, err := groupIdsFromId(id)
	if err != nil {
		return err
	}

	// Query main group
	group, resp, err := client.Groups.GetGroup(groupId, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] gitlab group %s not found so removing from state", groupId)
			d.SetId("")
			return nil
		}
		return err
	}

	// Find shared group data from queried group
	for _, sharedGroup := range group.SharedWithGroups {
		if sharedGroupId == sharedGroup.GroupID {
			convertedAccessLevel := gitlab.AccessLevelValue(sharedGroup.GroupAccessLevel)

			d.Set("group_id", groupId)
			d.Set("share_group_id", sharedGroup.GroupID)
			d.Set("group_access", accessLevel[convertedAccessLevel])

			if sharedGroup.ExpiresAt == nil {
				d.Set("expires_at", "")
			} else {
				d.Set("expires_at", sharedGroup.ExpiresAt.String())
			}

			return nil
		}
	}

	log.Printf("[DEBUG] gitlab shared group %s not found so removing from state", id)
	d.SetId("")
	return nil
}

func resourceGitlabGroupShareGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	id := d.Id()

	groupId, sharedGroupId, err := groupIdsFromId(id)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Delete gitlab share group %d for %s", sharedGroupId, groupId)

	_, err = client.GroupMembers.DeleteShareWithGroup(groupId, sharedGroupId)
	return err
}

func groupIdsFromId(id string) (string, int, error) {
	groupId, sharedGroupIdString, err := parseTwoPartID(id)
	if err != nil {
		return "", 0, fmt.Errorf("Error parsing ID: %s", id)
	}

	sharedGroupId, err := strconv.Atoi(sharedGroupIdString)
	if err != nil {
		return "", 0, fmt.Errorf("Can not determine shared group id: %s", sharedGroupIdString)
	}

	return groupId, sharedGroupId, nil
}
