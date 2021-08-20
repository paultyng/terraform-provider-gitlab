package gitlab

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
	"testing"
)

func TestAccGitlabManagedLicense_basic(t *testing.T) {
	var managedLicense gitlab.ManagedLicense
	rInt := acctest.RandInt()

	client := testAccNewClient(t)
	testAccCheckEE(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() {},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testManagedLicenseConfig(rInt, "approved"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabManagedLicenseExists("gitlab_managed_license.fix_me", &managedLicense),
				),
			},
			{
				Config: testManagedLicenseConfig(rInt, "blacklisted"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabManagedLicenseExists("gitlab_managed_license.fix_me", &managedLicense),
				),
			},
		},
	})
}

func testAccCheckGitlabManagedLicenseExists(n string, license *gitlab.ManagedLicense) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		licenseName := rs.Primary.ID
		project := rs.Primary.Attributes["project"]
		if project == "" {
			return fmt.Errorf("no project ID is set")
		}

		conn := testAccProvider.Meta().(*gitlab.Client)

		licenses, _, err := conn.ManagedLicenses.ListManagedLicenses(project)
		if err != nil {
			return err
		}

		for _, gotLicense := range licenses {
			if gotLicense.Name == licenseName {
				*license = *gotLicense
				return nil
			}
		}
		return fmt.Errorf("managed license does not exist")
	}
}

func testManagedLicenseConfig(rInt int, status string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_managed_license" "fixme" {
  project = "${gitlab_project.foo.id}"
  name = "FIXME-%d"
  approval_status = "%s"
}
	`, rInt, rInt, status)
}
