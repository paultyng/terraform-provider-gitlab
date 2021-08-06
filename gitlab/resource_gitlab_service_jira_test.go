package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	gitlab "github.com/xanzy/go-gitlab"
)

func TestAccGitlabServiceJira_basic(t *testing.T) {
	var jiraService gitlab.JiraService
	rInt := acctest.RandInt()
	jiraResourceName := "gitlab_service_jira.jira"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabServiceJiraDestroy,
		Steps: []resource.TestStep{
			// Create a project and a jira service
			{
				Config: testAccGitlabServiceJiraConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceJiraExists(jiraResourceName, &jiraService),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://test.com"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "user1"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "mypass"),
					resource.TestCheckResourceAttr(jiraResourceName, "project_key", "PROJ"),
					resource.TestCheckResourceAttr(jiraResourceName, "commit_events", "true"),
					resource.TestCheckResourceAttr(jiraResourceName, "merge_requests_events", "false"),
					resource.TestCheckResourceAttr(jiraResourceName, "comment_on_event_enabled", "false"),
				),
			},
			// Update the jira service
			{
				Config: testAccGitlabServiceJiraUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceJiraExists(jiraResourceName, &jiraService),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://testurl.com"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "user2"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "mypass_update"),
					resource.TestCheckResourceAttr(jiraResourceName, "project_key", "PROJ"),
					resource.TestCheckResourceAttr(jiraResourceName, "jira_issue_transition_id", "3"),
					resource.TestCheckResourceAttr(jiraResourceName, "commit_events", "false"),
					resource.TestCheckResourceAttr(jiraResourceName, "merge_requests_events", "true"),
					resource.TestCheckResourceAttr(jiraResourceName, "comment_on_event_enabled", "true"),
				),
			},
			// Update the jira service to get back to previous settings
			{
				Config: testAccGitlabServiceJiraConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabServiceJiraExists(jiraResourceName, &jiraService),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://test.com"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "user1"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "mypass"),
					resource.TestCheckResourceAttr(jiraResourceName, "project_key", "PROJ"),
					resource.TestCheckResourceAttr(jiraResourceName, "commit_events", "true"),
					resource.TestCheckResourceAttr(jiraResourceName, "merge_requests_events", "false"),
					resource.TestCheckResourceAttr(jiraResourceName, "comment_on_event_enabled", "false"),
				),
			},
		},
	})
}

// lintignore: AT002 // TODO: Resolve this tfproviderlint issue
func TestAccGitlabServiceJira_import(t *testing.T) {
	jiraResourceName := "gitlab_service_jira.jira"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabServiceJiraDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabServiceJiraConfig(rInt),
			},
			{
				ResourceName:      jiraResourceName,
				ImportStateIdFunc: getJiraProjectID(jiraResourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func testAccCheckGitlabServiceJiraExists(n string, service *gitlab.JiraService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		project := rs.Primary.Attributes["project"]
		if project == "" {
			return fmt.Errorf("No project ID is set")
		}
		conn := testAccProvider.Meta().(*gitlab.Client)

		jiraService, _, err := conn.Services.GetJiraService(project)
		if err != nil {
			return fmt.Errorf("Jira service does not exist in project %s: %v", project, err)
		}
		*service = *jiraService

		return nil
	}
}

func testAccCheckGitlabServiceJiraDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_project" {
			continue
		}

		gotRepo, resp, err := conn.Projects.GetProject(rs.Primary.ID, nil)
		if err == nil {
			if gotRepo != nil && fmt.Sprintf("%d", gotRepo.ID) == rs.Primary.ID {
				if gotRepo.MarkedForDeletionAt == nil {
					return fmt.Errorf("Repository still exists")
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

func getJiraProjectID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", n)
		}

		project := rs.Primary.Attributes["project"]
		if project == "" {
			return "", fmt.Errorf("No project ID is set")
		}

		return project, nil
	}
}

func testAccGitlabServiceJiraConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"
  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_service_jira" "jira" {
  project  = "${gitlab_project.foo.id}"
  url      = "https://test.com"
	username = "user1"
	project_key = "PROJ"
  password = "mypass"
  commit_events = true
  merge_requests_events    = false
  comment_on_event_enabled = false
}
`, rInt)
}

func testAccGitlabServiceJiraUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"
  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_service_jira" "jira" {
  project  = "${gitlab_project.foo.id}"
  url      = "https://testurl.com"
  username = "user2"
	password = "mypass_update"
	project_key = "PROJ"
  jira_issue_transition_id = "3"
  commit_events = false
  merge_requests_events    = true
  comment_on_event_enabled = true
}
`, rInt)
}
