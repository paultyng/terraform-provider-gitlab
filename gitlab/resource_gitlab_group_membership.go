package gitlab

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/xanzy/go-gitlab"
)

func resourceGitlabGroupMembership() *schema.Resource {
	acceptedAccessLevels := make([]string, 0, len(accessLevelID))
	for k := range accessLevelID {
		acceptedAccessLevels = append(acceptedAccessLevels, k)
	}
	return &schema.Resource{
		Create: resourceGitlabGroupMembershipCreate,
		Read:   resourceGitlabGroupMembershipRead,
		Update: resourceGitlabGroupMembershipUpdate,
		Delete: resourceGitlabGroupMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"user_id": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Required: true,
			},
			"access_level": {
				Type:         schema.TypeString,
				ValidateFunc: validateValueFunc(acceptedAccessLevels),
				Required:     true,
			},
			"expires_at": {
				Type:         schema.TypeString, // Format YYYY-MM-DD
				ValidateFunc: validateDateFunc,
				Optional:     true,
			},
		},
	}
}

func resourceGitlabGroupMembershipCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	userId := d.Get("user_id").(int)
	groupId := d.Get("group_id").(string)
	expiresAt := d.Get("expires_at").(string)
	accessLevelId := accessLevelID[d.Get("access_level").(string)]

	options := &gitlab.AddGroupMemberOptions{
		UserID:      &userId,
		AccessLevel: &accessLevelId,
		ExpiresAt:   &expiresAt,
	}
	log.Printf("[DEBUG] create gitlab group groupMember for %d in %s", options.UserID, groupId)

	_, resp, err := client.GroupMembers.AddGroupMember(groupId, options)
	if err != nil {
		user, _, err := client.Users.CurrentUser()
		if err != nil {
			return err
		}

		// The user that creates the group is always added automatically as member
		if resp != nil && resp.StatusCode == http.StatusConflict && user.ID == userId {
			options := gitlab.EditGroupMemberOptions{
				AccessLevel: &accessLevelId,
				ExpiresAt:   &expiresAt,
			}
			log.Printf("[DEBUG] update gitlab group membership %v for %s", userId, groupId)

			_, _, err := client.GroupMembers.EditGroupMember(groupId, userId, &options)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	userIdString := strconv.Itoa(userId)
	d.SetId(buildTwoPartID(&groupId, &userIdString))

	return resourceGitlabGroupMembershipRead(d, meta)
}

func resourceGitlabGroupMembershipRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	id := d.Id()
	log.Printf("[DEBUG] read gitlab group groupMember %s", id)

	groupId, userId, e := groupIdAndUserIdFromId(id)
	if e != nil {
		return e
	}

	groupMember, resp, err := client.GroupMembers.GetGroupMember(groupId, userId)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] gitlab group membership for %s not found so removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	resourceGitlabGroupMembershipSetToState(d, groupMember, &groupId)
	return nil
}

func groupIdAndUserIdFromId(id string) (string, int, error) {
	groupId, userIdString, err := parseTwoPartID(id)
	userId, e := strconv.Atoi(userIdString)
	if err != nil {
		e = err
	}
	if e != nil {
		log.Printf("[WARN] cannot get group member id from input: %v", id)
	}
	return groupId, userId, e
}

func resourceGitlabGroupMembershipUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	userId := d.Get("user_id").(int)
	groupId := d.Get("group_id").(string)
	expiresAt := d.Get("expires_at").(string)
	accessLevelId := accessLevelID[strings.ToLower(d.Get("access_level").(string))]

	options := gitlab.EditGroupMemberOptions{
		AccessLevel: &accessLevelId,
		ExpiresAt:   &expiresAt,
	}
	log.Printf("[DEBUG] update gitlab group membership %v for %s", userId, groupId)

	_, _, err := client.GroupMembers.EditGroupMember(groupId, userId, &options)
	if err != nil {
		return err
	}

	return resourceGitlabGroupMembershipRead(d, meta)
}

func resourceGitlabGroupMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	id := d.Id()
	groupId, userId, e := groupIdAndUserIdFromId(id)
	if e != nil {
		return e
	}

	log.Printf("[DEBUG] Delete gitlab group membership %v for %s", userId, groupId)

	_, err := client.GroupMembers.RemoveGroupMember(groupId, userId)
	return err
}

func resourceGitlabGroupMembershipSetToState(d *schema.ResourceData, groupMember *gitlab.GroupMember, groupId *string) {

	d.Set("group_id", groupId)
	d.Set("user_id", groupMember.ID)
	d.Set("access_level", accessLevel[groupMember.AccessLevel])
	d.Set("expires_at", groupMember.ExpiresAt) // lintignore: R004,XR004 // TODO: Resolve this tfproviderlint issue

	userId := strconv.Itoa(groupMember.ID)
	d.SetId(buildTwoPartID(groupId, &userId))
}
