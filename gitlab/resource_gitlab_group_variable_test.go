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
						Key:   fmt.Sprintf("key_%s", rString),
						Value: fmt.Sprintf("value-%s", rString),
					}),
				),
			},
			// Update the group variable to toggle all the values to their inverse
			{
				Config: testAccGitlabGroupVariableUpdateConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupVariableExists("gitlab_group_variable.foo", &groupVariable),
					testAccCheckGitlabGroupVariableAttributes(&groupVariable, &testAccGitlabGroupVariableExpectedAttributes{
						Key:       fmt.Sprintf("key_%s", rString),
						Value:     fmt.Sprintf("value-inverse-%s", rString),
						Protected: true,
					}),
				),
			},
			// Update the group variable to toggle the options back
			{
				Config: testAccGitlabGroupVariableConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupVariableExists("gitlab_group_variable.foo", &groupVariable),
					testAccCheckGitlabGroupVariableAttributes(&groupVariable, &testAccGitlabGroupVariableExpectedAttributes{
						Key:       fmt.Sprintf("key_%s", rString),
						Value:     fmt.Sprintf("value-%s", rString),
						Protected: false,
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
	Key       string
	Value     string
	Protected bool
	Masked    bool
}

func testAccCheckGitlabGroupVariableAttributes(variable *gitlab.GroupVariable, want *testAccGitlabGroupVariableExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != want.Key {
			return fmt.Errorf("got key %s; want %s", variable.Key, want.Key)
		}

		if variable.Value != want.Value {
			return fmt.Errorf("got value %s; value %s", variable.Value, want.Value)
		}

		if variable.Protected != want.Protected {
			return fmt.Errorf("got protected %t; want %t", variable.Protected, want.Protected)
		}

		if variable.Masked != want.Masked {
			return fmt.Errorf("got masked %t; want %t", variable.Masked, want.Masked)
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

		_, resp, err := conn.Groups.GetGroup(rs.Primary.ID, nil)
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
