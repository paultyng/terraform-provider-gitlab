package gitlab

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	gitlab "github.com/xanzy/go-gitlab"
)

func TestAccGitlabDeployKeyEnable_basic(t *testing.T) {
	var deployKey gitlab.DeployKey
	rInt := acctest.RandInt()

	keyTitle := "main"
	key := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDblguSWgpqiXIjHPSas4+N3Dten7MTLJMlGQXxGpaqN9nGPdNmuRB2YXyjT/nrryoY/qrtuVkPnis5WVo8N/s3hAnJbeJPUS2WKEGjpBlL34AQ+ANnlmGY8L6zr82Hp2Ommb7XGGtlq5D3yLCgTfcXLjC51tgcdwHsdH1U+RisgLwaTSrP/HF4G7IAr5ATsyYjtCwQRQ8ijdf5A34+XN6h8J6TLXKab5eZDuH38s9LxJuS7MRxx/P2UTOsqfjtrZWoQgE5adEGvnDxKyruex9PzNbCNVahzsma7tdikDbzxlHLIZ1aht6rKuai3iyLgcZfGIYtkq4xvg/bnNXxSsGf worker@kg.getwifi.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabDeployKeyEnableDestroy,
		Steps: []resource.TestStep{
			// Enable a deployKey on project with default options
			{
				Config: testAccGitlabDeployKeyEnableConfig(rInt, keyTitle, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabDeployKeyEnableExists("gitlab_deploy_key_enable.foo", &deployKey),
					testAccCheckGitlabDeployKeyEnableAttributes(&deployKey, &testAccGitlabDeployKeyEnableExpectedAttributes{
						Title:   keyTitle,
						Key:     key,
						CanPush: false,
					}),
				),
			},
			// Define canPush to true
			{
				Config: testAccGitlabDeployKeyEnableConfigCanPush(rInt, keyTitle, key, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabDeployKeyEnableExists("gitlab_deploy_key_enable.foo", &deployKey),
					testAccCheckGitlabDeployKeyEnableAttributes(&deployKey, &testAccGitlabDeployKeyEnableExpectedAttributes{
						Title:   keyTitle,
						Key:     key,
						CanPush: true,
					}),
				),
			},
			// Define canPush to false
			{
				Config: testAccGitlabDeployKeyEnableConfigCanPush(rInt, keyTitle, key, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabDeployKeyEnableExists("gitlab_deploy_key_enable.foo", &deployKey),
					testAccCheckGitlabDeployKeyEnableAttributes(&deployKey, &testAccGitlabDeployKeyEnableExpectedAttributes{
						Title:   keyTitle,
						Key:     key,
						CanPush: false,
					}),
				),
			},
			// Get back to default options
			{
				Config: testAccGitlabDeployKeyEnableConfig(rInt, keyTitle, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabDeployKeyEnableExists("gitlab_deploy_key_enable.foo", &deployKey),
					testAccCheckGitlabDeployKeyEnableAttributes(&deployKey, &testAccGitlabDeployKeyEnableExpectedAttributes{
						Title:   keyTitle,
						Key:     key,
						CanPush: false,
					}),
				),
			},
		},
	})
}

func testAccCheckGitlabDeployKeyEnableExists(n string, deployKey *gitlab.DeployKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		deployKeyID, err := strconv.Atoi(strings.Split(rs.Primary.ID, ":")[1])
		if err != nil {
			return err
		}
		repoName := rs.Primary.Attributes["project"]
		if repoName == "" {
			return fmt.Errorf("No project ID is set")
		}
		conn := testAccProvider.Meta().(*gitlab.Client)

		gotDeployKey, _, err := conn.DeployKeys.GetDeployKey(repoName, deployKeyID)
		if err != nil {
			return err
		}
		*deployKey = *gotDeployKey
		return nil
	}
}

type testAccGitlabDeployKeyEnableExpectedAttributes struct {
	Title   string
	Key     string
	CanPush bool
}

func testAccCheckGitlabDeployKeyEnableAttributes(deployKey *gitlab.DeployKey, want *testAccGitlabDeployKeyEnableExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if deployKey.Title != want.Title {
			return fmt.Errorf("got title %q; want %q", deployKey.Title, want.Title)
		}

		if deployKey.Key != want.Key {
			return fmt.Errorf("got key %q; want %q", deployKey.Key, want.Key)
		}

		if deployKey.CanPush != nil && *deployKey.CanPush != want.CanPush {
			return fmt.Errorf("got can_push %t; want %t", *deployKey.CanPush, want.CanPush)
		}

		return nil
	}
}

func testAccCheckGitlabDeployKeyEnableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	var project string
	var deployKeyID int

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "gitlab_project" {
			project = rs.Primary.Attributes["project"]
		} else if rs.Type == "gitlab_deploy_key_enable" {
			deployKeyID, _ = strconv.Atoi(strings.Split(rs.Primary.ID, ":")[1])
		}
	}

	gotDeployKey, resp, err := conn.DeployKeys.GetDeployKey(project, deployKeyID)
	if err == nil {
		if gotDeployKey != nil {
			return fmt.Errorf("Deploy key still exists: %d", deployKeyID)
		}
	}
	if resp.StatusCode != 404 {
		return err
	}
	return nil
}

func testAccGitlabDeployKeyEnableConfig(rInt int, keyTitle string, key string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "parent" {
  name = "parent-%d"
  description = "Terraform acceptance tests - Parent project"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests - Test Project"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_deploy_key" "parent" {
  project = "${gitlab_project.parent.id}"
  title = "%s"
  key = "%s"
}

resource "gitlab_deploy_key_enable" "foo" {
  project = "${gitlab_project.foo.id}"
  key_id = "${gitlab_deploy_key.parent.id}"
}
  `, rInt, rInt, keyTitle, key)
}

func testAccGitlabDeployKeyEnableConfigCanPush(rInt int, keyTitle string, key string, canPush bool) string {
	return fmt.Sprintf(`
resource "gitlab_project" "parent" {
  name = "parent-%d"
  description = "Terraform acceptance tests - Parent project"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests - Test Project"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_deploy_key" "parent" {
  project = "${gitlab_project.parent.id}"
  title = "%s"
  key = "%s"
}

resource "gitlab_deploy_key_enable" "foo" {
  project = "${gitlab_project.foo.id}"
  key_id = "${gitlab_deploy_key.parent.id}"
  can_push = %t
}
  `, rInt, rInt, keyTitle, key, canPush)
}
