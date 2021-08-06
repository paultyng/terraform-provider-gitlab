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

func TestAccGitlabProjectEnvironment_basic(t *testing.T) {
	rInt := acctest.RandInt()

	var env gitlab.Environment = gitlab.Environment{
		Name: fmt.Sprintf("ProjectEnvironment-%d", rInt),
	}
	var env2 gitlab.Environment = gitlab.Environment{
		Name:        fmt.Sprintf("ProjectEnvironment-%d", rInt),
		ExternalURL: "https://example.com",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectEnvironmentDestroy,
		Steps: []resource.TestStep{
			// Create a project and Environment with default options
			{
				Config: testAccGitlabProjectEnvironmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectEnvironmentExists("gitlab_project_environment.ProjectEnvironment", &env),
					testAccCheckGitlabProjectEnvironmentAttributes(&env, &testAccGitlabProjectEnvironmentExpectedAttributes{
						Name: fmt.Sprintf("ProjectEnvironment-%d", rInt),
					}),
				),
			},
			// Update the Environment
			{
				Config: testAccGitlabProjectEnvironmentUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectEnvironmentExists("gitlab_project_environment.ProjectEnvironment", &env),
					testAccCheckGitlabProjectEnvironmentAttributes(&env2, &testAccGitlabProjectEnvironmentExpectedAttributes{
						Name:        fmt.Sprintf("ProjectEnvironment-%d", rInt),
						ExternalURL: "https://example.com",
					}),
				),
			},
			// Update the Environment to get back to initial settings
			{
				Config: testAccGitlabProjectEnvironmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectEnvironmentExists("gitlab_project_environment.ProjectEnvironment", &env),
					testAccCheckGitlabProjectEnvironmentAttributes(&env, &testAccGitlabProjectEnvironmentExpectedAttributes{
						Name: fmt.Sprintf("ProjectEnvironment-%d", rInt),
					}),
				),
			},
		},
	})
}

func TestAccGitlabProjectEnvironment_wildcard(t *testing.T) {
	rInt := acctest.RandInt()

	var env gitlab.Environment = gitlab.Environment{
		Name: fmt.Sprintf("ProjectEnvironment-%d", rInt),
	}

	var env2 gitlab.Environment = gitlab.Environment{
		Name:        fmt.Sprintf("ProjectEnvironment-%d", rInt),
		ExternalURL: "https://example.com",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectEnvironmentDestroy,
		Steps: []resource.TestStep{
			// Create a project and Environment with default options
			{
				Config: testAccGitlabProjectEnvironmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectEnvironmentExists("gitlab_project_environment.ProjectEnvironment", &env),
					testAccCheckGitlabProjectEnvironmentAttributes(&env, &testAccGitlabProjectEnvironmentExpectedAttributes{
						Name: fmt.Sprintf("ProjectEnvironment-%d", rInt),
					}),
				),
			},
			// Update the Environment
			{
				Config: testAccGitlabProjectEnvironmentUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectEnvironmentExists("gitlab_project_environment.ProjectEnvironment", &env),
					testAccCheckGitlabProjectEnvironmentAttributes(&env2, &testAccGitlabProjectEnvironmentExpectedAttributes{
						Name:        fmt.Sprintf("ProjectEnvironment-%d", rInt),
						ExternalURL: "https://example.com",
					}),
				),
			},
			// Update the Environment to get back to initial settings
			{
				Config: testAccGitlabProjectEnvironmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectEnvironmentExists("gitlab_project_environment.ProjectEnvironment", &env),
					testAccCheckGitlabProjectEnvironmentAttributes(&env, &testAccGitlabProjectEnvironmentExpectedAttributes{
						Name: fmt.Sprintf("ProjectEnvironment-%d", rInt),
					}),
				),
			},
		},
	})
}

func testAccCheckGitlabProjectEnvironmentExists(n string, env *gitlab.Environment) resource.TestCheckFunc {
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

		environmentID, err := strconv.Atoi(environment)
		if err != nil {
			return fmt.Errorf("error converting environment ID to int: %v", err)
		}

		if _, _, err := conn.Environments.GetEnvironment(project, environmentID); err != nil {
			return err
		}
		return nil
	}
}

type testAccGitlabProjectEnvironmentExpectedAttributes struct {
	Name        string
	ExternalURL string
}

func testAccCheckGitlabProjectEnvironmentAttributes(env *gitlab.Environment, want *testAccGitlabProjectEnvironmentExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if env.Name != want.Name {
			return fmt.Errorf("got name %q; want %q", env.Name, want.Name)
		}

		if env.ExternalURL != want.ExternalURL {
			return fmt.Errorf("got external URL %q; want %q", env.ExternalURL, want.ExternalURL)
		}

		return nil
	}
}

func testAccCheckGitlabProjectEnvironmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)
	var project string
	var environment int
	var err error
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "gitlab_project" {
			project = rs.Primary.ID
		} else if rs.Type == "gitlab_project_environment" {
			environment, err = strconv.Atoi(rs.Primary.ID)
		}
	}

	env, response, err := conn.Environments.GetEnvironment(project, environment)
	if err == nil {
		if env != nil {
			return fmt.Errorf("project Environment %v still exists", environment)
		}
	}
	if response != nil && response.StatusCode != 404 {
		return err
	}
	return nil
}

func testAccGitlabProjectEnvironmentConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%[1]d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level       = "public"
}

resource "gitlab_project_environment" "ProjectEnvironment" {
  project = gitlab_project.foo.id
  name    = "ProjectEnvironment-%[1]d"
}
`, rInt)
}

func testAccGitlabProjectEnvironmentUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%[1]d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level       = "public"
}

resource "gitlab_project_environment" "ProjectEnvironment" {
  project      = gitlab_project.foo.id
  name         = "ProjectEnvironment-%[1]d"
  external_url = "https://example.com"
}
`, rInt)
}
