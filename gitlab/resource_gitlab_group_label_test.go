package gitlab

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabGroupLabel_basic(t *testing.T) {
	var label gitlab.GroupLabel
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabGroupLabelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabGroupLabelConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme", &label),
					testAccCheckGitlabGroupLabelAttributes(&label, &testAccGitlabGroupLabelExpectedAttributes{
						Name:        fmt.Sprintf("FIXME-%d", rInt),
						Color:       "#ffcc00",
						Description: "fix this test",
					}),
				),
			},
			{
				Config: testAccGitlabGroupLabelUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme", &label),
					testAccCheckGitlabGroupLabelAttributes(&label, &testAccGitlabGroupLabelExpectedAttributes{
						Name:        fmt.Sprintf("FIXME-%d", rInt),
						Color:       "#ff0000",
						Description: "red label",
					}),
				),
			},
			{
				Config: testAccGitlabGroupLabelConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme", &label),
					testAccCheckGitlabGroupLabelAttributes(&label, &testAccGitlabGroupLabelExpectedAttributes{
						Name:        fmt.Sprintf("FIXME-%d", rInt),
						Color:       "#ffcc00",
						Description: "fix this test",
					}),
				),
			},
			{
				Config: testAccGitlabGroupLabelLargeConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme.20", &label),
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme.30", &label),
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme.40", &label),
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme.10", &label),
					testAccCheckGitlabGroupLabelAttributes(&label, &testAccGitlabGroupLabelExpectedAttributes{
						Name:        "FIXME11",
						Color:       "#ffcc00",
						Description: "fix this test",
					}),
				),
			},
			{
				Config: testAccGitlabGroupLabelUpdateLargeConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme.20", &label),
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme.30", &label),
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme.40", &label),
					testAccCheckGitlabGroupLabelExists("gitlab_group_label.fixme.10", &label),
					testAccCheckGitlabGroupLabelAttributes(&label, &testAccGitlabGroupLabelExpectedAttributes{
						Name:        "FIXME11",
						Color:       "#ff0000",
						Description: "red label",
					}),
				),
			},
		},
	})
}

// lintignore: AT002 // TODO: Resolve this tfproviderlint issue
func TestAccGitlabGroupLabel_import(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "gitlab_group_label.fixme"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabGroupLabelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabGroupLabelConfig(rInt),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: getGroupLabelImportID(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func getGroupLabelImportID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", n)
		}

		labelID := rs.Primary.ID
		if labelID == "" {
			return "", fmt.Errorf("No deploy key ID is set")
		}
		groupID := rs.Primary.Attributes["group"]
		if groupID == "" {
			return "", fmt.Errorf("No group ID is set")
		}

		return fmt.Sprintf("%s:%s", groupID, labelID), nil
	}
}

func testAccCheckGitlabGroupLabelExists(n string, label *gitlab.GroupLabel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		labelName := rs.Primary.ID
		groupName := rs.Primary.Attributes["group"]
		if groupName == "" {
			return fmt.Errorf("No group ID is set")
		}
		conn := testAccProvider.Meta().(*gitlab.Client)

		labels, _, err := conn.GroupLabels.ListGroupLabels(groupName, &gitlab.ListGroupLabelsOptions{PerPage: 1000})
		if err != nil {
			return err
		}
		for _, gotLabel := range labels {
			if gotLabel.Name == labelName {
				*label = *gotLabel
				return nil
			}
		}
		return fmt.Errorf("Label does not exist")
	}
}

type testAccGitlabGroupLabelExpectedAttributes struct {
	Name        string
	Color       string
	Description string
}

func testAccCheckGitlabGroupLabelAttributes(label *gitlab.GroupLabel, want *testAccGitlabGroupLabelExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if label.Name != want.Name {
			return fmt.Errorf("got name %q; want %q", label.Name, want.Name)
		}

		if label.Description != want.Description {
			return fmt.Errorf("got description %q; want %q", label.Description, want.Description)
		}

		if label.Color != want.Color {
			return fmt.Errorf("got color %q; want %q", label.Color, want.Color)
		}

		return nil
	}
}

func testAccCheckGitlabGroupLabelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_group" {
			continue
		}

		group, resp, err := conn.Groups.GetGroup(rs.Primary.ID, nil)
		if err == nil {
			if group != nil && fmt.Sprintf("%d", group.ID) == rs.Primary.ID {
				if group.MarkedForDeletionOn == nil {
					return fmt.Errorf("Group still exists")
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

func testAccGitlabGroupLabelConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name             = "foo-%d"
  path             = "foo-%d"
  description      = "Terraform acceptance tests"
  visibility_level = "public"
}

resource "gitlab_group_label" "fixme" {
  group       = "${gitlab_group.foo.id}"
  name        = "FIXME-%d"
  color       = "#ffcc00"
  description = "fix this test"
}
	`, rInt, rInt, rInt)
}

func testAccGitlabGroupLabelUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name             = "foo-%d"
  path             = "foo-%d"
  description      = "Terraform acceptance tests"
  visibility_level = "public"
}

resource "gitlab_group_label" "fixme" {
  group       = "${gitlab_group.foo.id}"
  name        = "FIXME-%d"
  color       = "#ff0000"
  description = "red label"
}
	`, rInt, rInt, rInt)
}

func testAccGitlabGroupLabelLargeConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name             = "foo-%d"
  path             = "foo-%d"
  description      = "Terraform acceptance tests"
  visibility_level = "public"
}

resource "gitlab_group_label" "fixme" {
  group = "${gitlab_group.foo.id}"
  name = format("FIXME%%02d", count.index+1)
  count = 100
  color = "#ffcc00"
  description = "fix this test"
}
	`, rInt, rInt)
}

func testAccGitlabGroupLabelUpdateLargeConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name             = "foo-%d"
  path             = "foo-%d"
  description      = "Terraform acceptance tests"
  visibility_level = "public"
}

resource "gitlab_group_label" "fixme" {
  group = "${gitlab_group.foo.id}"
  name = format("FIXME%%02d", count.index+1)
  count = 99
  color = "#ff0000"
  description = "red label"
}
	`, rInt, rInt)
}
