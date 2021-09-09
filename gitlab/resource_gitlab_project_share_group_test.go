package gitlab

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabProjectShareGroup_basic(t *testing.T) {
	randName := acctest.RandomWithPrefix("acctest")

	// lintignore: AT001 // TODO: Resolve this tfproviderlint issue
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Share a new project with a new group.
			{
				Config: testAccGitlabProjectShareGroupConfig(randName, "guest"),
				Check:  testAccCheckGitlabProjectSharedWithGroup("root/"+randName, randName, gitlab.GuestPermissions),
			},
			// Update the access level.
			{
				Config: testAccGitlabProjectShareGroupConfig(randName, "reporter"),
				Check:  testAccCheckGitlabProjectSharedWithGroup("root/"+randName, randName, gitlab.ReporterPermissions),
			},
			// Delete the gitlab_project_share_group resource.
			{
				Config: testAccGitlabProjectShareGroupConfigDeleteShare(randName),
				Check:  testAccCheckGitlabProjectIsNotShared("root/" + randName),
			},
		},
	})
}

func testAccCheckGitlabProjectSharedWithGroup(projectName, groupName string, accessLevel gitlab.AccessLevelValue) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccProvider.Meta().(*gitlab.Client)

		project, _, err := client.Projects.GetProject(projectName, nil)
		if err != nil {
			return err
		}

		group, _, err := client.Groups.GetGroup(groupName, nil)
		if err != nil {
			return err
		}

		for _, share := range project.SharedWithGroups {
			if share.GroupID == group.ID {
				if gitlab.AccessLevelValue(share.GroupAccessLevel) != accessLevel {
					return fmt.Errorf("groupAccessLevel was %d (wanted %d)", share.GroupAccessLevel, accessLevel)
				}
				return nil
			}
		}

		return errors.New("project is not shared with group")
	}
}

func testAccCheckGitlabProjectIsNotShared(projectName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client := testAccProvider.Meta().(*gitlab.Client)

		project, _, err := client.Projects.GetProject(projectName, nil)
		if err != nil {
			return err
		}

		if len(project.SharedWithGroups) != 0 {
			return fmt.Errorf("project is shared with %d groups (wanted 0)", len(project.SharedWithGroups))
		}

		return nil
	}
}

func testAccGitlabProjectShareGroupConfig(randName, accessLevel string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "test" {
  name = "%[1]s"

  # So that acceptance tests can be run in a gitlab organization with no billing.
  visibility_level = "public"
}

resource "gitlab_group" "test" {
  name = "%[1]s"
  path = "%[1]s"
}

resource "gitlab_project_share_group" "test" {
  project_id = gitlab_project.test.id
  group_id = gitlab_group.test.id
  access_level = "%[2]s"
}
`, randName, accessLevel)
}

func testAccGitlabProjectShareGroupConfigDeleteShare(randName string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "test" {
  name = "%[1]s"

  # So that acceptance tests can be run in a gitlab organization with no billing.
  visibility_level = "public"
}

resource "gitlab_group" "test" {
  name = "%[1]s"
  path = "%[1]s"
}
`, randName)
}
