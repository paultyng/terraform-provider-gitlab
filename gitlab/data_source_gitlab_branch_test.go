package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataGitlabBranch_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataGitlabBranch(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceGitlabBranch("gitlab_branch.foo", "data.gitlab_branch.foo"),
				),
			},
		},
	})
}

func testAccDataSourceGitlabBranch(src, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		branch := s.RootModule().Resources[src]
		branchAttr := branch.Primary.Attributes

		search := s.RootModule().Resources[n]
		searchAttr := search.Primary.Attributes

		testAttributes := []string{
			"id",
			"name",
			"web_url",
			"default",
			"project",
			"can_push",
			"merged",
			"commit",
			"parent_ids",
		}

		for _, attribute := range testAttributes {
			if searchAttr[attribute] != branchAttr[attribute] {
				return fmt.Errorf("expected branch's parameter `%s` to be: %s, but got: `%s`", attribute, branchAttr[attribute], searchAttr[attribute])
			}
		}
		return nil
	}
}

func testAccDataGitlabBranch(rInt int) string {
	return fmt.Sprintf(`
%s

data "gitlab_branch" "foo" {
  name = "${gitlab_branch.foo.name}"
  project = "${gitlab_branch.foo.project}"
}
`, testAccDataGitlabBranchSetup(rInt))
}

func testAccDataGitlabBranchSetup(rInt int) string {
	return fmt.Sprintf(`
	resource "gitlab_project" "test" {
		name = "foo-%d"
		description = "Terraform acceptance tests"
	  
		# So that acceptance tests can be run in a gitlab organization
		# with no billing
		visibility_level = "public"
	}
	resource "gitlab_branch" "foo" {
		name = "testbranch-%d"
		ref = "main"
		project = gitlab_project.test.id
	}
  `, rInt, rInt)
}
