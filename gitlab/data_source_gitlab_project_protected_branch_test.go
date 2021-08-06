package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataGitlabProjectProtectedBranchSearch(t *testing.T) {
	projectName := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataGitlabProjectProtectedBranchConfigGetProjectSearch(projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.gitlab_project_protected_branch.test",
						"name",
						"master",
					),
					resource.TestCheckResourceAttr(
						"data.gitlab_project_protected_branch.test",
						"push_access_levels.0.access_level",
						"maintainer",
					),
				),
			},
		},
	})
}

func testAccDataGitlabProjectProtectedBranchConfigGetProjectSearch(projectName string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "test" {
  name           = "%s"
  path           = "%s"
  default_branch = "master"
}

resource "gitlab_branch_protection" "master" {
  project            = gitlab_project.test.id
  branch             = "master"
  push_access_level  = "maintainer"
  merge_access_level = "developer"
}

resource "gitlab_branch_protection" "test" {
  project            = gitlab_project.test.id
  branch             = "master"
  push_access_level  = "maintainer"
  merge_access_level = "developer"
}

data "gitlab_project_protected_branch" "test" {
  project_id = gitlab_project.test.id
  name       = gitlab_branch_protection.master.branch
}
`, projectName, projectName)
}
