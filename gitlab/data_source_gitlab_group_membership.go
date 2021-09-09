package gitlab

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/xanzy/go-gitlab"
)

func dataSourceGitlabGroupMembership() *schema.Resource {
	acceptedAccessLevels := make([]string, 0, len(accessLevelID))
	for k := range accessLevelID {
		acceptedAccessLevels = append(acceptedAccessLevels, k)
	}
	return &schema.Resource{
		Read: dataSourceGitlabGroupMembershipRead,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
				ConflictsWith: []string{
					"full_path",
				},
			},
			"full_path": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ConflictsWith: []string{
					"group_id",
				},
			},
			"access_level": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validateValueFunc(acceptedAccessLevels),
			},
			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"avatar_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"web_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"access_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expires_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGitlabGroupMembershipRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	var gm []*gitlab.GroupMember
	var group *gitlab.Group
	var err error

	log.Printf("[INFO] Reading Gitlab group")

	groupIDData, groupIDOk := d.GetOk("group_id")
	fullPathData, fullPathOk := d.GetOk("full_path")

	if groupIDOk {
		// Get group by id
		group, _, err = client.Groups.GetGroup(groupIDData.(int), nil)
		if err != nil {
			return err
		}
	} else if fullPathOk {
		// Get group by full path
		group, _, err = client.Groups.GetGroup(fullPathData.(string), nil)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("one and only one of group_id or full_path must be set")
	}

	log.Printf("[INFO] Reading Gitlab group memberships")

	// Get group memberships
	gm, _, err = client.Groups.ListGroupMembers(group.ID, &gitlab.ListGroupMembersOptions{})
	if err != nil {
		return err
	}

	d.Set("group_id", group.ID)
	d.Set("full_path", group.FullPath)

	d.Set("members", flattenGitlabMembers(d, gm)) // lintignore: XR004 // TODO: Resolve this tfproviderlint issue

	var optionsHash strings.Builder
	optionsHash.WriteString(strconv.Itoa(group.ID))

	if data, ok := d.GetOk("access_level"); ok {
		optionsHash.WriteString(data.(string))
	}

	id := schema.HashString(optionsHash.String())
	d.SetId(fmt.Sprintf("%d", id))

	return nil
}

func flattenGitlabMembers(d *schema.ResourceData, members []*gitlab.GroupMember) []interface{} {
	membersList := []interface{}{}

	var filterAccessLevel gitlab.AccessLevelValue = gitlab.NoPermissions
	if data, ok := d.GetOk("access_level"); ok {
		filterAccessLevel = accessLevelID[data.(string)]
	}

	for _, member := range members {
		if filterAccessLevel != gitlab.NoPermissions && filterAccessLevel != member.AccessLevel {
			continue
		}

		values := map[string]interface{}{
			"id":           member.ID,
			"username":     member.Username,
			"name":         member.Name,
			"state":        member.State,
			"avatar_url":   member.AvatarURL,
			"web_url":      member.WebURL,
			"access_level": accessLevel[gitlab.AccessLevelValue(member.AccessLevel)],
		}

		if member.ExpiresAt != nil {
			values["expires_at"] = member.ExpiresAt.String()
		}

		membersList = append(membersList, values)
	}

	return membersList
}
