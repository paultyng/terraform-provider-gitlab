package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabGroupVariable_basic(t *testing.T) {
	var groupVariable gitlab.GroupVariable
	rString := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabGroupVariableDestroy,
		Steps: []resource.TestStep{
			// Create a group and variable with default options
			{
				Config: testAccGitlabGroupVariableConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupVariableExists("gitlab_group_variable.foo", &groupVariable),
					testAccCheckGitlabGroupVariableAttributes(&groupVariable, &testAccGitlabGroupVariableExpectedAttributes{
						Key:              fmt.Sprintf("key_%s", rString),
						Value:            fmt.Sprintf("value-%s", rString),
						EnvironmentScope: "*",
					}),
				),
			},
			// Update the group variable to toggle all the values to their inverse - check environment_scope if license allows it
			{
				Config:   testAccGitlabGroupVariableUpdateConfigWithEnvironmentScope(rString),
				SkipFunc: isRunningInCE,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupVariableExists("gitlab_group_variable.foo", &groupVariable),
					testAccCheckGitlabGroupVariableAttributes(&groupVariable, &testAccGitlabGroupVariableExpectedAttributes{
						Key:              fmt.Sprintf("key_%s", rString),
						Value:            fmt.Sprintf("value-inverse-%s", rString),
						Protected:        true,
						EnvironmentScope: fmt.Sprintf("foo%s", rString),
					}),
				),
			},
			// Update the group variable to toggle all the values to their inverse - skip check of environment_scope as it is only available in Premium
			{
				Config:   testAccGitlabGroupVariableUpdateConfig(rString),
				SkipFunc: isRunningInEE,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupVariableExists("gitlab_group_variable.foo", &groupVariable),
					testAccCheckGitlabGroupVariableAttributes(&groupVariable, &testAccGitlabGroupVariableExpectedAttributes{
						Key:              fmt.Sprintf("key_%s", rString),
						Value:            fmt.Sprintf("value-inverse-%s", rString),
						Protected:        true,
						EnvironmentScope: "*",
					}),
				),
			},
			// Update the group variable to toggle the options back
			{
				Config: testAccGitlabGroupVariableConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupVariableExists("gitlab_group_variable.foo", &groupVariable),
					testAccCheckGitlabGroupVariableAttributes(&groupVariable, &testAccGitlabGroupVariableExpectedAttributes{
						Key:              fmt.Sprintf("key_%s", rString),
						Value:            fmt.Sprintf("value-%s", rString),
						Protected:        false,
						EnvironmentScope: "*",
					}),
				),
			},
		},
	})
}

func testAccCheckGitlabGroupVariableExists(n string, groupVariable *gitlab.GroupVariable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		repoName := rs.Primary.Attributes["group"]
		if repoName == "" {
			return fmt.Errorf("No group ID is set")
		}
		key := rs.Primary.Attributes["key"]
		if key == "" {
			return fmt.Errorf("No variable key is set")
		}
		conn := testAccProvider.Meta().(*gitlab.Client)

		gotVariable, _, err := conn.GroupVariables.GetVariable(repoName, key)
		if err != nil {
			return err
		}
		*groupVariable = *gotVariable
		return nil
	}
}

type testAccGitlabGroupVariableExpectedAttributes struct {
	Key              string
	Value            string
	Protected        bool
	Masked           bool
	EnvironmentScope string
}

func testAccCheckGitlabGroupVariableAttributes(variable *gitlab.GroupVariable, want *testAccGitlabGroupVariableExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != want.Key {
			return fmt.Errorf("got key %q; want %q", variable.Key, want.Key)
		}

		if variable.Value != want.Value {
			return fmt.Errorf("got value %q; value %q", variable.Value, want.Value)
		}

		if variable.Protected != want.Protected {
			return fmt.Errorf("got protected %t; want %t", variable.Protected, want.Protected)
		}

		if variable.Masked != want.Masked {
			return fmt.Errorf("got masked %t; want %t", variable.Masked, want.Masked)
		}

		if variable.EnvironmentScope != want.EnvironmentScope {
			return fmt.Errorf("got environment_scope %q; want %q", variable.EnvironmentScope, want.EnvironmentScope)
		}

		return nil
	}
}

func testAccCheckGitlabGroupVariableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_group" {
			continue
		}

		_, resp, err := conn.Groups.GetGroup(rs.Primary.ID)
		if err == nil { // nolint // TODO: Resolve this golangci-lint issue: SA9003: empty branch (staticcheck)
			//if gotRepo != nil && fmt.Sprintf("%d", gotRepo.ID) == rs.Primary.ID {
			//	if gotRepo.MarkedForDeletionAt == nil {
			//		return fmt.Errorf("Repository still exists")
			//	}
			//}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabGroupVariableConfig(rString string) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
name = "foo%v"
path = "foo%v"
}

resource "gitlab_group_variable" "foo" {
  group = "${gitlab_group.foo.id}"
  key = "key_%s"
  value = "value-%s"
  variable_type = "file"
  masked = false
}
	`, rString, rString, rString, rString)
}

func testAccGitlabGroupVariableUpdateConfig(rString string) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
name = "foo%v"
path = "foo%v"
}

resource "gitlab_group_variable" "foo" {
  group = "${gitlab_group.foo.id}"
  key = "key_%s"
  value = "value-inverse-%s"
  protected = true
  masked = false
}
	`, rString, rString, rString, rString)
}

func testAccGitlabGroupVariableUpdateConfigWithEnvironmentScope(rString string) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
name = "foo%v"
path = "foo%v"
}

resource "gitlab_group_variable" "foo" {
  group = "${gitlab_group.foo.id}"
  key = "key_%s"
  value = "value-inverse-%s"
  protected = true
  masked = false
  environment_scope = "foo%s"
}
	`, rString, rString, rString, rString, rString)
}
