package gitlab

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabGroupBadge_basic(t *testing.T) {
	var badge gitlab.GroupBadge
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabGroupBadgeDestroy,
		Steps: []resource.TestStep{
			// Create a group and badge
			{
				Config: testAccGitlabGroupBadgeConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupBadgeExists("gitlab_group_badge.foo", &badge),
					testAccCheckGitlabGroupBadgeAttributes(&badge, &testAccGitlabGroupBadgeExpectedAttributes{
						LinkURL:  fmt.Sprintf("http://example.com/badge-%d", rInt),
						ImageURL: fmt.Sprintf("http://example.com/badge-%d.svg", rInt),
					}),
				),
			},
			// Test ImportState
			{
				ResourceName:      "gitlab_group_badge.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update the group badge
			{
				Config: testAccGitlabGroupBadgeUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupBadgeExists("gitlab_group_badge.foo", &badge),
					testAccCheckGitlabGroupBadgeAttributes(&badge, &testAccGitlabGroupBadgeExpectedAttributes{
						LinkURL:  fmt.Sprintf("https://example.com/badge-%d", rInt),
						ImageURL: fmt.Sprintf("https://example.com/badge-%d.svg", rInt),
					}),
				),
			},
		},
	})
}

func testAccCheckGitlabGroupBadgeExists(n string, badge *gitlab.GroupBadge) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		splitID := strings.Split(rs.Primary.ID, ":")

		badgeID, err := strconv.Atoi(splitID[len(splitID)-1])
		if err != nil {
			return err
		}
		groupID := rs.Primary.Attributes["group"]
		if groupID == "" {
			return fmt.Errorf("No group ID is set")
		}

		conn := testAccProvider.Meta().(*gitlab.Client)

		gotBadge, _, err := conn.GroupBadges.GetGroupBadge(groupID, badgeID)
		if err != nil {
			return err
		}
		*badge = *gotBadge
		return nil
	}
}

type testAccGitlabGroupBadgeExpectedAttributes struct {
	LinkURL  string
	ImageURL string
}

func testAccCheckGitlabGroupBadgeAttributes(badge *gitlab.GroupBadge, want *testAccGitlabGroupBadgeExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if badge.LinkURL != want.LinkURL {
			return fmt.Errorf("got link_url %q; want %q", badge.LinkURL, want.LinkURL)
		}

		if badge.ImageURL != want.ImageURL {
			return fmt.Errorf("got image_url %s; want %s", badge.ImageURL, want.ImageURL)
		}

		return nil
	}
}

func testAccCheckGitlabGroupBadgeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_group" {
			continue
		}

		group, resp, err := conn.Groups.GetGroup(rs.Primary.ID, nil)
		if err == nil {
			if group != nil && fmt.Sprintf("%d", group.ID) == rs.Primary.ID {
				if group.MarkedForDeletionOn == nil {
					return fmt.Errorf("Group still exists")
				}
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabGroupBadgeConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name        = "foo-%d"
  path        = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_group_badge" "foo" {
  group     = "${gitlab_group.foo.id}"
  link_url  = "http://example.com/badge-%d"
  image_url = "http://example.com/badge-%d.svg"
}
	`, rInt, rInt, rInt, rInt)
}

func testAccGitlabGroupBadgeUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name        = "foo-%d"
  path        = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

# change schema from http to https
resource "gitlab_group_badge" "foo" {
  group     = "${gitlab_group.foo.id}"
  link_url  = "https://example.com/badge-%d"
  image_url = "https://example.com/badge-%d.svg"
}
	`, rInt, rInt, rInt, rInt)
}
