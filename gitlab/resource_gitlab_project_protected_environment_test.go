package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabProjectProtectedEnvironment_basic(t *testing.T) {

	var pt gitlab.ProtectedEnvironment
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectProtectedEnvironmentDestroy,
		Steps: []resource.TestStep{
			// Create a project and Protected Environment with default options
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabProjectProtectedEnvironmentConfig(rInt, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectProtectedEnvironmentExists("gitlab_project_protected_environment.ProjectProtectedEnvironment", &pt),
					testAccCheckGitlabProjectProtectedEnvironmentAttributes(&pt, &testAccGitlabProjectProtectedEnvironmentExpectedAttributes{
						Name:              fmt.Sprintf("ProjectProtectedEnvironment-%d", rInt),
						CreateAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
			// Update the Protected Environment
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabProjectProtectedEnvironmentUpdateConfig(rInt, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectProtectedEnvironmentExists("gitlab_project_protected_environment.ProjectProtectedEnvironment", &pt),
					testAccCheckGitlabProjectProtectedEnvironmentAttributes(&pt, &testAccGitlabProjectProtectedEnvironmentExpectedAttributes{
						Name:              fmt.Sprintf("ProjectProtectedEnvironment-%d", rInt),
						CreateAccessLevel: accessLevel[gitlab.MasterPermissions],
					}),
				),
			},
			// Update the Protected Environment to get back to initial settings
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabProjectProtectedEnvironmentConfig(rInt, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectProtectedEnvironmentExists("gitlab_project_protected_environment.ProjectProtectedEnvironment", &pt),
					testAccCheckGitlabProjectProtectedEnvironmentAttributes(&pt, &testAccGitlabProjectProtectedEnvironmentExpectedAttributes{
						Name:              fmt.Sprintf("ProjectProtectedEnvironment-%d", rInt),
						CreateAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
		},
	})
}

func TestAccGitlabProjectProtectedEnvironment_wildcard(t *testing.T) {

	var pt gitlab.ProtectedEnvironment
	rInt := acctest.RandInt()

	wildcard := "-*"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectProtectedEnvironmentDestroy,
		Steps: []resource.TestStep{
			// Create a project and Protected Environment with default options
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabProjectProtectedEnvironmentConfig(rInt, wildcard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectProtectedEnvironmentExists("gitlab_project_protected_environment.ProjectProtectedEnvironment", &pt),
					testAccCheckGitlabProjectProtectedEnvironmentAttributes(&pt, &testAccGitlabProjectProtectedEnvironmentExpectedAttributes{
						Name:              fmt.Sprintf("ProjectProtectedEnvironment-%d%s", rInt, wildcard),
						CreateAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
			// Update the Protected Environment
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabProjectProtectedEnvironmentUpdateConfig(rInt, wildcard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectProtectedEnvironmentExists("gitlab_project_protected_environment.ProjectProtectedEnvironment", &pt),
					testAccCheckGitlabProjectProtectedEnvironmentAttributes(&pt, &testAccGitlabProjectProtectedEnvironmentExpectedAttributes{
						Name:              fmt.Sprintf("ProjectProtectedEnvironment-%d%s", rInt, wildcard),
						CreateAccessLevel: accessLevel[gitlab.MasterPermissions],
					}),
				),
			},
			// Update the Protected Environment to get back to initial settings
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabProjectProtectedEnvironmentConfig(rInt, wildcard),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectProtectedEnvironmentExists("gitlab_project_protected_environment.ProjectProtectedEnvironment", &pt),
					testAccCheckGitlabProjectProtectedEnvironmentAttributes(&pt, &testAccGitlabProjectProtectedEnvironmentExpectedAttributes{
						Name:              fmt.Sprintf("ProjectProtectedEnvironment-%d%s", rInt, wildcard),
						CreateAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
		},
	})
}

func testAccCheckGitlabProjectProtectedEnvironmentExists(n string, pt *gitlab.ProtectedEnvironment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}
		project, environment, err := parseTwoPartID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error in Splitting Project ID and Environment Name")
		}

		conn := testAccProvider.Meta().(*gitlab.Client)

		pes, _, err := conn.ProtectedEnvironments.ListProtectedEnvironments(project, nil)
		if err != nil {
			return err
		}
		for _, gotpe := range pes {
			if gotpe.Name == environment {
				*pt = *gotpe
				return nil
			}
		}
		return fmt.Errorf("Protected Environment does not exist")
	}
}

type testAccGitlabProjectProtectedEnvironmentExpectedAttributes struct {
	Name              string
	CreateAccessLevel string
}

func testAccCheckGitlabProjectProtectedEnvironmentAttributes(pt *gitlab.ProtectedEnvironment, want *testAccGitlabProjectProtectedEnvironmentExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if pt.Name != want.Name {
			return fmt.Errorf("got name %q; want %q", pt.Name, want.Name)
		}

		if pt.DeployAccessLevels[0].AccessLevel != accessLevelID[want.CreateAccessLevel] {
			return fmt.Errorf("got Create access levels %q; want %q", pt.DeployAccessLevels[0].AccessLevel, accessLevelID[want.CreateAccessLevel])
		}

		return nil
	}
}

func testAccCheckGitlabProjectProtectedEnvironmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)
	var project string
	var environment string
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "gitlab_project" {
			project = rs.Primary.ID
		} else if rs.Type == "gitlab_project_protected_environment" {
			environment = rs.Primary.ID
		}
	}

	pt, response, err := conn.ProtectedEnvironments.GetProtectedEnvironment(project, environment)
	if err == nil {
		if pt != nil {
			return fmt.Errorf("project Protected Environment %s still exists", environment)
		}
	}
	if response != nil && response.StatusCode != 404 {
		return err
	}
	return nil
}

func testAccGitlabProjectProtectedEnvironmentConfig(rInt int, postfix string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%[1]d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level       = "public"
  shared_runners_enabled = true
}

resource "gitlab_project_environment" "env" {
  project = gitlab_project.foo.id
  name    = "ProjectProtectedEnvironment-%[1]d-this-suffix-matches-wildcard"
}

resource "gitlab_project_protected_environment" "ProjectProtectedEnvironment" {
  depends_on  = [gitlab_project_environment.env]
  project     = gitlab_project.foo.id
  environment = "ProjectProtectedEnvironment-%[1]d%[2]s"
  deploy_access_levels {
	access_level = "developer"
  }
}
	`, rInt, postfix)
}

func testAccGitlabProjectProtectedEnvironmentUpdateConfig(rInt int, postfix string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%[1]d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level       = "public"
  shared_runners_enabled = true
}

resource "gitlab_project_environment" "env" {
  project = gitlab_project.foo.id
  name    = "ProjectProtectedEnvironment-%[1]d-this-suffix-matches-wildcard"
}

resource "gitlab_user" "test" {
  name           = "test-%[1]d"
  username       = "test-%[1]d"
  password       = "superPassword-%[1]d"
  email          = "test-%[1]d@example.com"
  reset_password = false
}

resource "gitlab_group" "test" {
  name = "test-%[1]d"
  path = "test-%[1]d"
}

resource "gitlab_project_protected_environment" "ProjectProtectedEnvironment" {
  depends_on  = [gitlab_project_environment.env]
  project     = gitlab_project.foo.id
  environment = "ProjectProtectedEnvironment-%[1]d%[2]s"
  deploy_access_levels {
    access_level = "maintainer"
  }
  deploy_access_levels {
    user_id = gitlab_user.test.id
  }
  deploy_access_levels {
    group_id = gitlab_group.test.id
  }
}
	`, rInt, postfix)
}
