package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataGitlabProjectProtectedBranchesSearch(t *testing.T) {
	projectName := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataGitlabProjectProtectedBranchesConfigGetProjectSearch(projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.gitlab_project_protected_branches.test",
						"protected_branches.0.name",
						"master",
					),
					resource.TestCheckResourceAttr(
						"data.gitlab_project_protected_branches.test",
						"protected_branches.0.push_access_levels.0.access_level",
						"maintainer",
					),
				),
			},
		},
	})
}

func testAccDataGitlabProjectProtectedBranchesConfigGetProjectSearch(projectName string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "test" {
  name           = "%s"
  path           = "%s"
  default_branch = "master"
}

resource "gitlab_branch_protection" "test" {
  project            = gitlab_project.test.id
  branch             = "master"
  push_access_level  = "maintainer"
  merge_access_level = "developer"
}

data "gitlab_project_protected_branches" "test" {
  project_id = gitlab_branch_protection.test.project # This expresses the dependency of the data source on the protected branch having first been configured
}
`, projectName, projectName)
}
