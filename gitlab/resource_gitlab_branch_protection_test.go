package gitlab

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabBranchProtection_basic(t *testing.T) {

	var pb gitlab.ProtectedBranch
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabBranchProtectionDestroy,
		Steps: []resource.TestStep{
			// Create a project and Branch Protection with default options
			{
				Config: testAccGitlabBranchProtectionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionComputedAttributes("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
			// Update the Branch Protection
			{
				Config: testAccGitlabBranchProtectionUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.MasterPermissions],
						MergeAccessLevel: accessLevel[gitlab.MasterPermissions],
					}),
				),
			},
			// Update the Branch Protection to get back to initial settings
			{
				Config: testAccGitlabBranchProtectionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
			// Update the the Branch Protection code owner approval setting
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabBranchProtectionUpdateConfigCodeOwnerTrue(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:                      fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:           accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel:          accessLevel[gitlab.DeveloperPermissions],
						CodeOwnerApprovalRequired: true,
					}),
				),
			},
			// Attempting to update code owner approval setting on CE should fail safely and with an informative error message
			{
				SkipFunc:    isRunningInEE,
				Config:      testAccGitlabBranchProtectionUpdateConfigCodeOwnerTrue(rInt),
				ExpectError: regexp.MustCompile("feature unavailable: code owner approvals"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
			// Update the Branch Protection to get back to initial settings
			{
				Config: testAccGitlabBranchProtectionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
		},
	})
}

func TestAccGitlabBranchProtection_createWithCodeOwnerApproval(t *testing.T) {
	var pb gitlab.ProtectedBranch
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabBranchProtectionDestroy,
		Steps: []resource.TestStep{
			// Start with code owner approval required disabled
			{
				SkipFunc: isRunningInEE,
				Config:   testAccGitlabBranchProtectionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
			// Create a project and Branch Protection with code owner approval enabled
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabBranchProtectionUpdateConfigCodeOwnerTrue(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:                      fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:           accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel:          accessLevel[gitlab.DeveloperPermissions],
						CodeOwnerApprovalRequired: true,
					}),
				),
			},
			// Attempting to update code owner approval setting on CE should fail safely and with an informative error message
			{
				SkipFunc:    isRunningInEE,
				Config:      testAccGitlabBranchProtectionUpdateConfigCodeOwnerTrue(rInt),
				ExpectError: regexp.MustCompile("feature unavailable: code owner approvals"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:                      fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:           accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel:          accessLevel[gitlab.DeveloperPermissions],
						CodeOwnerApprovalRequired: true,
					}),
				),
			},
			// Update the Branch Protection to get back to initial settings
			{
				Config: testAccGitlabBranchProtectionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:             fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:  accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel: accessLevel[gitlab.DeveloperPermissions],
					}),
				),
			},
		},
	})
}

func TestAccGitlabBranchProtection_createWithMultipleAccessLevels(t *testing.T) {
	var pb gitlab.ProtectedBranch
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabBranchProtectionDestroy,
		Steps: []resource.TestStep{
			// Create a project, groups, users and Branch Protection with advanced allowed_to blocks
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabBranchProtectionConfigMultipleAccessLevels(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:                 fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:      accessLevel[gitlab.MaintainerPermissions],
						MergeAccessLevel:     accessLevel[gitlab.MaintainerPermissions],
						UsersAllowedToPush:   []string{fmt.Sprintf("listest2%d", rInt)},
						UsersAllowedToMerge:  []string{fmt.Sprintf("listest%d", rInt), fmt.Sprintf("listest2%d", rInt)},
						GroupsAllowedToPush:  []string{fmt.Sprintf("test-%d", rInt), fmt.Sprintf("test2-%d", rInt)},
						GroupsAllowedToMerge: []string{fmt.Sprintf("test-%d", rInt), fmt.Sprintf("test2-%d", rInt)},
					}),
				),
			},
			// Update to remove some allowed_to blocks and update access levels
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabBranchProtectionConfigMultipleAccessLevelsUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabBranchProtectionExists("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionPersistsInStateCorrectly("gitlab_branch_protection.branch_protect", &pb),
					testAccCheckGitlabBranchProtectionAttributes(&pb, &testAccGitlabBranchProtectionExpectedAttributes{
						Name:                 fmt.Sprintf("BranchProtect-%d", rInt),
						PushAccessLevel:      accessLevel[gitlab.DeveloperPermissions],
						MergeAccessLevel:     accessLevel[gitlab.DeveloperPermissions],
						UsersAllowedToPush:   []string{fmt.Sprintf("listest%d", rInt)},
						UsersAllowedToMerge:  []string{fmt.Sprintf("listest2%d", rInt)},
						GroupsAllowedToPush:  []string{fmt.Sprintf("test2-%d", rInt)},
						GroupsAllowedToMerge: []string{fmt.Sprintf("test-%d", rInt)},
					}),
				),
			},
		},
	})
}

func testAccCheckGitlabBranchProtectionPersistsInStateCorrectly(n string, pb *gitlab.ProtectedBranch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		var mergeAccessLevel gitlab.AccessLevelValue
		for _, v := range pb.MergeAccessLevels {
			if v.UserID == 0 && v.GroupID == 0 {
				mergeAccessLevel = v.AccessLevel
				break
			}
		}
		if rs.Primary.Attributes["merge_access_level"] != accessLevelValueToName[mergeAccessLevel] {
			return fmt.Errorf("merge access level not persisted in state correctly")
		}

		var pushAccessLevel gitlab.AccessLevelValue
		for _, v := range pb.PushAccessLevels {
			if v.UserID == 0 && v.GroupID == 0 {
				pushAccessLevel = v.AccessLevel
				break
			}
		}
		if rs.Primary.Attributes["push_access_level"] != accessLevelValueToName[pushAccessLevel] {
			return fmt.Errorf("push access level not persisted in state correctly")
		}

		if rs.Primary.Attributes["code_owner_approval_required"] != strconv.FormatBool(pb.CodeOwnerApprovalRequired) {
			return fmt.Errorf("code_owner_approval_required not persisted in state correctly")
		}

		return nil
	}
}

func testAccCheckGitlabBranchProtectionExists(n string, pb *gitlab.ProtectedBranch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}
		project, branch, err := projectAndBranchFromID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error in Splitting Project and Branch Ids")
		}

		conn := testAccProvider.Meta().(*gitlab.Client)

		pbs, _, err := conn.ProtectedBranches.ListProtectedBranches(project, nil)
		if err != nil {
			return err
		}
		for _, gotpb := range pbs {
			if gotpb.Name == branch {
				*pb = *gotpb
				return nil
			}
		}
		return fmt.Errorf("Protected Branch does not exist")
	}
}

func testAccCheckGitlabBranchProtectionComputedAttributes(n string, pb *gitlab.ProtectedBranch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(n, "branch_protection_id", strconv.Itoa(pb.ID))(s)
	}
}

type testAccGitlabBranchProtectionExpectedAttributes struct {
	Name                      string
	PushAccessLevel           string
	MergeAccessLevel          string
	UsersAllowedToPush        []string
	UsersAllowedToMerge       []string
	GroupsAllowedToPush       []string
	GroupsAllowedToMerge      []string
	CodeOwnerApprovalRequired bool
}

func testAccCheckGitlabBranchProtectionAttributes(pb *gitlab.ProtectedBranch, want *testAccGitlabBranchProtectionExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if pb.Name != want.Name {
			return fmt.Errorf("got name %q; want %q", pb.Name, want.Name)
		}

		var pushAccessLevel gitlab.AccessLevelValue
		for _, v := range pb.PushAccessLevels {
			if v.UserID == 0 && v.GroupID == 0 {
				pushAccessLevel = v.AccessLevel
				break
			}
		}
		if pushAccessLevel != accessLevelID[want.PushAccessLevel] {
			return fmt.Errorf("got Push access level %v; want %v", pushAccessLevel, accessLevelID[want.PushAccessLevel])
		}

		var mergeAccessLevel gitlab.AccessLevelValue
		for _, v := range pb.MergeAccessLevels {
			if v.UserID == 0 && v.GroupID == 0 {
				mergeAccessLevel = v.AccessLevel
				break
			}
		}
		if mergeAccessLevel != accessLevelID[want.MergeAccessLevel] {
			return fmt.Errorf("got Merge access level %v; want %v", mergeAccessLevel, accessLevelID[want.MergeAccessLevel])
		}

		conn := testAccProvider.Meta().(*gitlab.Client)

		remainingWantedUserIDsAllowedToPush := map[int]struct{}{}
		for _, v := range want.UsersAllowedToPush {
			users, _, err := conn.Users.ListUsers(&gitlab.ListUsersOptions{
				Username: gitlab.String(v),
			})
			if err != nil {
				return fmt.Errorf("error looking up user by path %v: %v", v, err)
			}
			if len(users) != 1 {
				return fmt.Errorf("error finding user by username %v; found %v", v, len(users))
			}
			remainingWantedUserIDsAllowedToPush[users[0].ID] = struct{}{}
		}
		remainingWantedGroupIDsAllowedToPush := map[int]struct{}{}
		for _, v := range want.GroupsAllowedToPush {
			group, _, err := conn.Groups.GetGroup(v, nil)
			if err != nil {
				return fmt.Errorf("error looking up group by path %v: %v", v, err)
			}
			remainingWantedGroupIDsAllowedToPush[group.ID] = struct{}{}
		}
		for _, v := range pb.PushAccessLevels {
			if v.UserID != 0 {
				if _, ok := remainingWantedUserIDsAllowedToPush[v.UserID]; !ok {
					return fmt.Errorf("found unwanted user ID %v", v.UserID)
				}
				delete(remainingWantedUserIDsAllowedToPush, v.UserID)
			} else if v.GroupID != 0 {
				if _, ok := remainingWantedGroupIDsAllowedToPush[v.GroupID]; !ok {
					return fmt.Errorf("found unwanted group ID %v", v.GroupID)
				}
				delete(remainingWantedGroupIDsAllowedToPush, v.GroupID)
			}
		}
		if len(remainingWantedUserIDsAllowedToPush) > 0 {
			return fmt.Errorf("failed to find wanted user IDs %v", remainingWantedUserIDsAllowedToPush)
		}
		if len(remainingWantedGroupIDsAllowedToPush) > 0 {
			return fmt.Errorf("failed to find wanted group IDs %v", remainingWantedGroupIDsAllowedToPush)
		}

		if pb.CodeOwnerApprovalRequired != want.CodeOwnerApprovalRequired {
			return fmt.Errorf("got code_owner_approval_required %v; want %v", pb.CodeOwnerApprovalRequired, want.CodeOwnerApprovalRequired)
		}

		return nil
	}
}

func testAccCheckGitlabBranchProtectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)
	var project string
	var branch string
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "gitlab_project" {
			project = rs.Primary.ID
		} else if rs.Type == "gitlab_branch_protection" {
			branch = rs.Primary.ID
		}
	}

	pb, response, err := conn.ProtectedBranches.GetProtectedBranch(project, branch)
	if err == nil {
		if pb != nil {
			return fmt.Errorf("project branch protection %s still exists", branch)
		}
	}
	if response.StatusCode != 404 {
		return err
	}
	return nil
}

func testAccGitlabBranchProtectionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%[1]d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_branch_protection" "branch_protect" {
  project            = gitlab_project.foo.id
  branch             = "BranchProtect-%[1]d"
  push_access_level  = "developer"
  merge_access_level = "developer"
}
	`, rInt)
}

func testAccGitlabBranchProtectionUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%[1]d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_branch_protection" "branch_protect" {
  project            = gitlab_project.foo.id
  branch             = "BranchProtect-%[1]d"
  push_access_level  = "maintainer"
  merge_access_level = "maintainer"
}
	`, rInt)
}

func testAccGitlabBranchProtectionUpdateConfigCodeOwnerTrue(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%[1]d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_branch_protection" "branch_protect" {
  project                      = gitlab_project.foo.id
  branch                       = "BranchProtect-%[1]d"
  push_access_level            = "developer"
  merge_access_level           = "developer"
  code_owner_approval_required = true
}
	`, rInt)
}

func testAccGitlabBranchProtectionConfigMultipleAccessLevels(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "test" {
  name = "test-%[1]d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_group" "test" {
  name = "test-%[1]d"
  path = "test-%[1]d"
}

resource "gitlab_group" "test2" {
  name = "test2-%[1]d"
  path = "test2-%[1]d"
}

resource "gitlab_user" "test" {
  name             = "foo %[1]d"
  username         = "listest%[1]d"
  password         = "test%[1]dtt"
  email            = "listest%[1]d@ssss.com"
  is_admin         = false
  projects_limit   = 0
  can_create_group = false
  is_external      = false
}

resource "gitlab_user" "test2" {
  name             = "foo2 %[1]d"
  username         = "listest2%[1]d"
  password         = "test2%[1]dtt"
  email            = "listest2%[1]d@ssss.com"
  is_admin         = false
  projects_limit   = 0
  can_create_group = false
  is_external      = false
}

resource "gitlab_project_share_group" "test" {
  project_id   = gitlab_project.test.id
  group_id     = gitlab_group.test.id
  access_level = "developer"
}

resource "gitlab_project_share_group" "test2" {
  project_id   = gitlab_project.test.id
  group_id     = gitlab_group.test2.id
  access_level = "developer"
}

resource "gitlab_project_membership" "test" {
  project_id   = gitlab_project.test.id
  user_id      = gitlab_user.test.id
  access_level = "developer"
}

resource "gitlab_project_membership" "test2" {
  project_id   = gitlab_project.test.id
  user_id      = gitlab_user.test2.id
  access_level = "developer"
}

resource "gitlab_group_membership" "test" {
  depends_on   = [gitlab_project_share_group.test]
  group_id     = gitlab_group.test.id
  user_id      = gitlab_user.test.id
  access_level = "developer"
}

resource "gitlab_group_membership" "test2" {
  depends_on   = [gitlab_project_share_group.test2]
  group_id     = gitlab_group.test2.id
  user_id      = gitlab_user.test2.id
  access_level = "developer"
}

resource "gitlab_branch_protection" "branch_protect" {
  depends_on = [
	gitlab_group_membership.test,
	gitlab_group_membership.test2,
	gitlab_project_membership.test,
	gitlab_project_membership.test2,
  ]
  project            = gitlab_project.test.id
  branch             = "BranchProtect-%[1]d"
  push_access_level  = "maintainer"
  merge_access_level = "maintainer"
  allowed_to_push {
    user_id = gitlab_user.test2.id
  }
  allowed_to_push {
    group_id = gitlab_group.test.id
  }
  allowed_to_push {
    group_id = gitlab_group.test2.id
  }
  allowed_to_merge {
    user_id = gitlab_user.test.id
  }
  allowed_to_merge {
    group_id = gitlab_group.test.id
  }
  allowed_to_merge {
    user_id = gitlab_user.test2.id
  }
  allowed_to_merge {
    group_id = gitlab_group.test2.id
  }
}
	`, rInt)
}

func testAccGitlabBranchProtectionConfigMultipleAccessLevelsUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "test" {
  name = "test-%[1]d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_group" "test" {
  name = "test-%[1]d"
  path = "test-%[1]d"
}

resource "gitlab_group" "test2" {
  name = "test2-%[1]d"
  path = "test2-%[1]d"
}

resource "gitlab_user" "test" {
  name             = "test %[1]d"
  username         = "listest%[1]d"
  password         = "test%[1]dtt"
  email            = "listest%[1]d@ssss.com"
  is_admin         = false
  projects_limit   = 0
  can_create_group = false
  is_external      = false
}

resource "gitlab_user" "test2" {
  name             = "test2 %[1]d"
  username         = "listest2%[1]d"
  password         = "test2%[1]dtt"
  email            = "listest2%[1]d@ssss.com"
  is_admin         = false
  projects_limit   = 0
  can_create_group = false
  is_external      = false
}

resource "gitlab_project_share_group" "test" {
  project_id   = gitlab_project.test.id
  group_id     = gitlab_group.test.id
  access_level = "developer"
}

resource "gitlab_project_share_group" "test2" {
  project_id   = gitlab_project.test.id
  group_id     = gitlab_group.test2.id
  access_level = "developer"
}

resource "gitlab_project_membership" "test" {
  project_id   = gitlab_project.test.id
  user_id      = gitlab_user.test.id
  access_level = "developer"
}

resource "gitlab_project_membership" "test2" {
  project_id   = gitlab_project.test.id
  user_id      = gitlab_user.test2.id
  access_level = "developer"
}

resource "gitlab_group_membership" "test" {
  depends_on   = [gitlab_project_share_group.test]
  group_id     = gitlab_group.test.id
  user_id      = gitlab_user.test.id
  access_level = "developer"
}

resource "gitlab_group_membership" "test2" {
  depends_on   = [gitlab_project_share_group.test2]
  group_id     = gitlab_group.test2.id
  user_id      = gitlab_user.test2.id
  access_level = "developer"
}

resource "gitlab_branch_protection" "branch_protect" {
  depends_on = [
	gitlab_group_membership.test,
	gitlab_group_membership.test2,
	gitlab_project_membership.test,
	gitlab_project_membership.test2,
  ]
  project            = gitlab_project.test.id
  branch             = "BranchProtect-%[1]d"
  push_access_level  = "developer"
  merge_access_level = "developer"
  allowed_to_push {
    user_id = gitlab_user.test.id
  }
  allowed_to_push {
    group_id = gitlab_group.test2.id
  }
  allowed_to_merge {
    user_id = gitlab_user.test2.id
  }
  allowed_to_merge {
    group_id = gitlab_group.test.id
  }
}
	`, rInt)
}
