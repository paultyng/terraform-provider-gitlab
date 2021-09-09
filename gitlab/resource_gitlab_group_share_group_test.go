package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabGroupShareGroup_basic(t *testing.T) {
	randName := acctest.RandomWithPrefix("acctest")

	// lintignore: AT001 // TODO: Resolve this tfproviderlint issue
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Share a new group with another group
			{
				Config: testAccGitlabGroupShareGroupConfig(
					randName,
					`
					group_access 	 = "guest"
					expires_at     = "2099-01-01"
					`,
				),
				Check: testAccCheckGitlabGroupSharedWithGroup(randName, "2099-01-01", gitlab.GuestPermissions),
			},
			// Update the share group
			{
				Config: testAccGitlabGroupShareGroupConfig(randName, `group_access = "reporter"`),
				Check:  testAccCheckGitlabGroupSharedWithGroup(randName, "", gitlab.ReporterPermissions),
			},
			// Delete the gitlab_group_share_group resource
			{
				Config: testAccGitlabGroupShareGroupConfigDelete(randName),
				Check:  testAccCheckGitlabGroupIsNotShared(randName),
			},
		},
	})
}

// lintignore: AT002 // TODO: Resolve this tfproviderlint issue
func TestAccGitlabGroupShareGroup_import(t *testing.T) {
	randName := acctest.RandomWithPrefix("acctest")

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckGitlabGroupDestroy,
		Steps: []resource.TestStep{
			{
				// create shared groups
				Config: testAccGitlabGroupShareGroupConfig(
					randName,
					`
					group_access 	 = "guest"
					expires_at     = "2099-03-03"
					`,
				),
				Check: testAccCheckGitlabGroupSharedWithGroup(randName, "2099-03-03", gitlab.GuestPermissions),
			},
			{
				// Verify Import
				ResourceName:      "gitlab_group_share_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGitlabGroupSharedWithGroup(
	groupName string,
	expireTime string,
	accessLevel gitlab.AccessLevelValue,
) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccProvider.Meta().(*gitlab.Client)

		mainGroup, _, err := client.Groups.GetGroup(fmt.Sprintf("%s_main", groupName), nil)
		if err != nil {
			return err
		}

		sharedGroupsCount := len(mainGroup.SharedWithGroups)
		if sharedGroupsCount != 1 {
			return fmt.Errorf("Number of shared groups was %d (wanted %d)", sharedGroupsCount, 1)
		}

		sharedGroup := mainGroup.SharedWithGroups[0]

		if sharedGroup.GroupName != fmt.Sprintf("%s_share", groupName) {
			return fmt.Errorf("group name was %s (wanted %s)", sharedGroup.GroupName, fmt.Sprintf("%s_share", groupName))
		}

		if gitlab.AccessLevelValue(sharedGroup.GroupAccessLevel) != accessLevel {
			return fmt.Errorf("groupAccessLevel was %d (wanted %d)", sharedGroup.GroupAccessLevel, accessLevel)
		}

		if sharedGroup.ExpiresAt == nil && expireTime != "" {
			return fmt.Errorf("expire time was nil (wanted %s)", expireTime)
		} else if sharedGroup.ExpiresAt != nil && sharedGroup.ExpiresAt.String() != expireTime {
			return fmt.Errorf("expire time was %s (wanted %s)", sharedGroup.ExpiresAt.String(), expireTime)
		}

		return nil
	}
}

func testAccCheckGitlabGroupIsNotShared(groupName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccProvider.Meta().(*gitlab.Client)

		mainGroup, _, err := client.Groups.GetGroup(fmt.Sprintf("%s_main", groupName), nil)
		if err != nil {
			return err
		}

		sharedGroupsCount := len(mainGroup.SharedWithGroups)
		if sharedGroupsCount != 0 {
			return fmt.Errorf("Number of shared groups was %d (wanted %d)", sharedGroupsCount, 0)
		}

		return nil
	}
}

func testAccGitlabGroupShareGroupConfig(
	randName string,
	shareGroupSettings string,
) string {
	return fmt.Sprintf(
		`
		resource "gitlab_group" "test_main" {
		  name = "%[1]s_main"
		  path = "%[1]s_main"
		}

		resource "gitlab_group" "test_share" {
		  name = "%[1]s_share"
		  path = "%[1]s_share"
		}

		resource "gitlab_group_share_group" "test" {
		  group_id       = gitlab_group.test_main.id
			share_group_id = gitlab_group.test_share.id
			%[2]s
		}
		`,
		randName,
		shareGroupSettings,
	)
}

func testAccGitlabGroupShareGroupConfigDelete(randName string) string {
	return fmt.Sprintf(
		`
		resource "gitlab_group" "test_main" {
		  name = "%[1]s_main"
		  path = "%[1]s_main"
		}

		resource "gitlab_group" "test_share" {
		  name = "%[1]s_share"
		  path = "%[1]s_share"
		}
		`,
		randName,
	)
}
