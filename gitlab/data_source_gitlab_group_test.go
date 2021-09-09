package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceGitlabGroup_basic(t *testing.T) {
	rString := fmt.Sprintf("%s", acctest.RandString(5)) // nolint // TODO: Resolve this golangci-lint issue: S1025: the argument is already a string, there's no need to use fmt.Sprintf (gosimple)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Get group using its ID
			{
				Config: testAccDataGitlabGroupByID(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceGitlabGroup("gitlab_group.foo", "data.gitlab_group.foo"),
				),
			},
			// Get group using its full path
			{
				Config: testAccDataGitlabGroupByFullPath(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceGitlabGroup("gitlab_group.sub_foo", "data.gitlab_group.sub_foo"),
				),
			},
		},
	})
}

func testAccDataSourceGitlabGroup(src, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		group := s.RootModule().Resources[src]
		groupResource := group.Primary.Attributes

		search := s.RootModule().Resources[n]
		searchResource := search.Primary.Attributes

		testAttributes := []string{
			"id",
			"full_path",
			"name",
			"full_name",
			"web_url",
			"path",
			"description",
			"lfs_enabled",
			"request_access_enabled",
			"visibility_level",
			"parent_id",
			"default_branch_protection",
		}

		for _, attribute := range testAttributes {
			if searchResource[attribute] != groupResource[attribute] {
				return fmt.Errorf("expected group's parameter `%s` to be: %s, but got: `%s`", attribute, groupResource[attribute], searchResource[attribute])
			}
		}

		return nil
	}
}

func testAccDataGitlabGroupByID(rString string) string {
	return fmt.Sprintf(`
%s

data "gitlab_group" "foo" {
  group_id = "${gitlab_group.foo.id}"
}
`, testAccDataGitlabGroupSetup(rString))
}

func testAccDataGitlabGroupByFullPath(rString string) string {
	return fmt.Sprintf(`
%s

data "gitlab_group" "sub_foo" {
  full_path = "${gitlab_group.foo.path}/${gitlab_group.sub_foo.path}"
}
`, testAccDataGitlabGroupSetup(rString))
}

func testAccDataGitlabGroupSetup(rString string) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name = "foo-name-%[1]s"
  path = "foo-path-%[1]s"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_group" "sub_foo" {
  name      = "sub-foo-name-%[1]s"
  path      = "sub-foo-path-%[1]s"
  parent_id = "${gitlab_group.foo.id}"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}
  `, rString)
}
