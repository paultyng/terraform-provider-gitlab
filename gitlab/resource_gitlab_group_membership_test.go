package gitlab

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabGroupMembership_basic(t *testing.T) {
	var groupMember gitlab.GroupMember
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{PreCheck: func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabGroupMembershipDestroy,
		Steps: []resource.TestStep{

			// Assign member to the group as a developer
			{
				Config: testAccGitlabGroupMembershipConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabGroupMembershipExists("gitlab_group_membership.foo", &groupMember), testAccCheckGitlabGroupMembershipAttributes(&groupMember, &testAccGitlabGroupMembershipExpectedAttributes{
					accessLevel: fmt.Sprintf("developer"), // nolint // TODO: Resolve this golangci-lint issue: S1039: unnecessary use of fmt.Sprintf (gosimple)
				})),
			},

			//Update the group member to change the access level (use testAccGitlabGroupMembershipUpdateConfig for Config)
			{
				Config: testAccGitlabGroupMembershipUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabGroupMembershipExists("gitlab_group_membership.foo", &groupMember), testAccCheckGitlabGroupMembershipAttributes(&groupMember, &testAccGitlabGroupMembershipExpectedAttributes{
					accessLevel: fmt.Sprintf("guest"),      // nolint // TODO: Resolve this golangci-lint issue: S1039: unnecessary use of fmt.Sprintf (gosimple)
					expiresAt:   fmt.Sprintf("2099-01-01"), // nolint // TODO: Resolve this golangci-lint issue: S1039: unnecessary use of fmt.Sprintf (gosimple)
				})),
			},

			// Update the group member to change the access level back
			{
				Config: testAccGitlabGroupMembershipConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabGroupMembershipExists("gitlab_group_membership.foo", &groupMember), testAccCheckGitlabGroupMembershipAttributes(&groupMember, &testAccGitlabGroupMembershipExpectedAttributes{
					accessLevel: fmt.Sprintf("developer"), // nolint // TODO: Resolve this golangci-lint issue: S1039: unnecessary use of fmt.Sprintf (gosimple)
				})),
			},

			// Add the same member again
			{
				Config: testAccGitlabGroupMembershipConfig(rInt),
				Check: resource.ComposeTestCheckFunc(testAccCheckGitlabGroupMembershipExists("gitlab_group_membership.foo", &groupMember), testAccCheckGitlabGroupMembershipAttributes(&groupMember, &testAccGitlabGroupMembershipExpectedAttributes{
					accessLevel: fmt.Sprintf("developer"),
				})),
			},
		},
	})
}

func testAccCheckGitlabGroupMembershipExists(n string, membership *gitlab.GroupMember) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(*gitlab.Client)
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		groupId := rs.Primary.Attributes["group_id"]
		if groupId == "" {
			return fmt.Errorf("No group ID is set")
		}

		userIdString := rs.Primary.Attributes["user_id"]
		userId, _ := strconv.Atoi(userIdString)
		if userIdString == "" {
			return fmt.Errorf("No user userId is set")
		}

		gotGroupMembership, _, err := conn.GroupMembers.GetGroupMember(groupId, userId)
		if err != nil {
			return err
		}

		*membership = *gotGroupMembership
		return nil
	}
}

type testAccGitlabGroupMembershipExpectedAttributes struct {
	accessLevel string
	expiresAt   string
}

func testAccCheckGitlabGroupMembershipAttributes(membership *gitlab.GroupMember, want *testAccGitlabGroupMembershipExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		accessLevelId, ok := accessLevel[membership.AccessLevel]
		if !ok {
			return fmt.Errorf("Invalid access level '%s'", accessLevelId)
		}
		if accessLevelId != want.accessLevel {
			return fmt.Errorf("got access level %s; want %s", accessLevelId, want.accessLevel)
		}
		return nil
	}
}

func testAccCheckGitlabGroupMembershipDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_group_membership" {
			continue
		}

		groupId := rs.Primary.Attributes["group_id"]
		userIdString := rs.Primary.Attributes["user_id"]

		// GetGroupMember needs int type for userIdString
		userId, err := strconv.Atoi(userIdString) // nolint // TODO: Resolve this golangci-lint issue: ineffectual assignment to err (ineffassign)
		groupMember, resp, err := conn.GroupMembers.GetGroupMember(groupId, userId)
		if err != nil {
			if groupMember != nil && fmt.Sprintf("%d", groupMember.AccessLevel) == rs.Primary.Attributes["accessLevel"] {
				return fmt.Errorf("Group still has member.")
			}
			return nil
		}

		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabGroupMembershipConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name = "foo%d"
  path = "foo%d"
}

resource "gitlab_user" "test" {
  name 		= "foo%d"
  username  = "listest%d"
  password  = "test%dtt"
  email 	= "listest%d@ssss.com"
}

resource "gitlab_group_membership" "foo" {
  group_id 		= "${gitlab_group.foo.id}"
  user_id 		= "${gitlab_user.test.id}"
  access_level 	= "developer"
}`, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccGitlabGroupMembershipUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name = "foo%d"
  path = "foo%d"
}

resource "gitlab_user" "test" {
  name 		= "foo%d"
  username 	= "listest%d"
  password 	= "test%dtt"
  email 	= "listest%d@ssss.com"
}

resource "gitlab_group_membership" "foo" {
  group_id 		= "${gitlab_group.foo.id}"
  user_id 		= "${gitlab_user.test.id}"
  expires_at    = "2099-01-01"
  access_level 	= "guest"
}`, rInt, rInt, rInt, rInt, rInt, rInt)
}
