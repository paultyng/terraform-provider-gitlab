package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
)

func TestAccGitlabGroupLdapLink_basic(t *testing.T) {
	var testLdapLink gitlab.LDAPGroupLink
	var ldapLink gitlab.LDAPGroupLink
	rInt := acctest.RandInt()
	testDataFile := "testdata/resource_gitlab_group_ldap_link.json"

	// PreCheck runs after Config so load test data here
	err := testAccLoadTestData(testDataFile, &testLdapLink)
	if err != nil {
		t.Fatalf("[ERROR] Failed to load test data: %s", err.Error())
	}

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabGroupLdapLinkDestroy,
		Steps: []resource.TestStep{

			// Create a group LDAP link as a developer (uses testAccGitlabGroupLdapLinkCreateConfig for Config)
			{
				SkipFunc: testAccGitlabGroupLdapLinkSkipFunc(testLdapLink.CN, testLdapLink.Provider),
				Config:   testAccGitlabGroupLdapLinkCreateConfig(rInt, &testLdapLink),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLdapLinkExists("gitlab_group_ldap_link.foo", &ldapLink),
					testAccCheckGitlabGroupLdapLinkAttributes(&ldapLink, &testAccGitlabGroupLdapLinkExpectedAttributes{
						accessLevel: fmt.Sprintf("developer"), // nolint // TODO: Resolve this golangci-lint issue: S1039: unnecessary use of fmt.Sprintf (gosimple)
					})),
			},

			// Update the group LDAP link to change the access level (uses testAccGitlabGroupLdapLinkUpdateConfig for Config)
			{
				SkipFunc: testAccGitlabGroupLdapLinkSkipFunc(testLdapLink.CN, testLdapLink.Provider),
				Config:   testAccGitlabGroupLdapLinkUpdateConfig(rInt, &testLdapLink),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLdapLinkExists("gitlab_group_ldap_link.foo", &ldapLink),
					testAccCheckGitlabGroupLdapLinkAttributes(&ldapLink, &testAccGitlabGroupLdapLinkExpectedAttributes{
						accessLevel: fmt.Sprintf("maintainer"), // nolint // TODO: Resolve this golangci-lint issue: S1039: unnecessary use of fmt.Sprintf (gosimple)
					})),
			},

			// Force create the same group LDAP link in a different resource (uses testAccGitlabGroupLdapLinkForceCreateConfig for Config)
			{
				SkipFunc: testAccGitlabGroupLdapLinkSkipFunc(testLdapLink.CN, testLdapLink.Provider),
				Config:   testAccGitlabGroupLdapLinkForceCreateConfig(rInt, &testLdapLink),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLdapLinkExists("gitlab_group_ldap_link.bar", &ldapLink),
					testAccCheckGitlabGroupLdapLinkAttributes(&ldapLink, &testAccGitlabGroupLdapLinkExpectedAttributes{
						accessLevel: fmt.Sprintf("developer"), // nolint // TODO: Resolve this golangci-lint issue: S1039: unnecessary use of fmt.Sprintf (gosimple)
					})),
			},
		},
	})
}

func testAccGitlabGroupLdapLinkSkipFunc(testCN string, testProvider string) func() (bool, error) {
	return func() (bool, error) {
		if testCN == "default" || testProvider == "default" {
			return true, nil
		}

		return isRunningInCE()
	}
}

func testAccCheckGitlabGroupLdapLinkExists(resourceName string, ldapLink *gitlab.LDAPGroupLink) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Clear the "found" LDAP link before checking for existence
		*ldapLink = gitlab.LDAPGroupLink{}

		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		err := testAccGetGitlabGroupLdapLink(ldapLink, resourceState)
		if err != nil {
			return err
		}

		return nil
	}
}

type testAccGitlabGroupLdapLinkExpectedAttributes struct {
	accessLevel string
}

func testAccCheckGitlabGroupLdapLinkAttributes(ldapLink *gitlab.LDAPGroupLink, want *testAccGitlabGroupLdapLinkExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		accessLevelId, ok := accessLevel[ldapLink.GroupAccess]
		if !ok {
			return fmt.Errorf("Invalid access level '%s'", accessLevelId)
		}
		if accessLevelId != want.accessLevel {
			return fmt.Errorf("Has access level %s; want %s", accessLevelId, want.accessLevel)
		}
		return nil
	}
}

func testAccCheckGitlabGroupLdapLinkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	// Can't check for links if the group is destroyed so make sure all groups are destroyed instead
	for _, resourceState := range s.RootModule().Resources {
		if resourceState.Type != "gitlab_group" {
			continue
		}

		group, resp, err := conn.Groups.GetGroup(resourceState.Primary.ID, nil)
		if err == nil {
			if group != nil && fmt.Sprintf("%d", group.ID) == resourceState.Primary.ID {
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

func testAccGetGitlabGroupLdapLink(ldapLink *gitlab.LDAPGroupLink, resourceState *terraform.ResourceState) error {
	conn := testAccProvider.Meta().(*gitlab.Client)

	groupId := resourceState.Primary.Attributes["group_id"]
	if groupId == "" {
		return fmt.Errorf("No group ID is set")
	}

	// Construct our desired LDAP Link from the config values
	desiredLdapLink := gitlab.LDAPGroupLink{
		CN:          resourceState.Primary.Attributes["cn"],
		GroupAccess: accessLevelNameToValue[resourceState.Primary.Attributes["access_level"]],
		Provider:    resourceState.Primary.Attributes["ldap_provider"],
	}

	desiredLdapLinkId := buildTwoPartID(&desiredLdapLink.Provider, &desiredLdapLink.CN)

	// Try to fetch all group links from GitLab
	currentLdapLinks, _, err := conn.Groups.ListGroupLDAPLinks(groupId, nil)
	if err != nil {
		// The read/GET API wasn't implemented in GitLab until version 12.8 (March 2020, well after the add and delete APIs).
		// If we 404, assume GitLab is at an older version and take things on faith.
		switch err.(type) { // nolint // TODO: Resolve this golangci-lint issue: S1034: assigning the result of this type assertion to a variable (switch err := err.(type)) could eliminate type assertions in switch cases (gosimple)
		case *gitlab.ErrorResponse:
			if err.(*gitlab.ErrorResponse).Response.StatusCode == 404 { // nolint // TODO: Resolve this golangci-lint issue: S1034(related information): could eliminate this type assertion (gosimple)
				// Do nothing
			} else {
				return err
			}
		default:
			return err
		}
	}

	// If we got here and don't have links, assume GitLab is below version 12.8 and skip the check
	if currentLdapLinks != nil {
		found := false

		// Check if the LDAP link exists in the returned list of links
		for _, currentLdapLink := range currentLdapLinks {
			if buildTwoPartID(&currentLdapLink.Provider, &currentLdapLink.CN) == desiredLdapLinkId {
				found = true
				*ldapLink = *currentLdapLink
				break
			}
		}

		if !found {
			return errors.New(fmt.Sprintf("LdapLink %s does not exist.", desiredLdapLinkId)) // nolint // TODO: Resolve this golangci-lint issue: S1028: should use fmt.Errorf(...) instead of errors.New(fmt.Sprintf(...)) (gosimple)
		}
	} else {
		*ldapLink = desiredLdapLink
	}

	return nil
}

func testAccLoadTestData(testdatafile string, ldapLink *gitlab.LDAPGroupLink) error {
	testLdapLinkBytes, err := ioutil.ReadFile(testdatafile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(testLdapLinkBytes, ldapLink)
	if err != nil {
		return err
	}

	return nil
}

func testAccGitlabGroupLdapLinkCreateConfig(rInt int, testLdapLink *gitlab.LDAPGroupLink) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
    name = "foo%d"
	path = "foo%d"
	description = "Terraform acceptance test - Group LDAP Links 1"
}

resource "gitlab_group_ldap_link" "foo" {
    group_id 		= "${gitlab_group.foo.id}"
    cn				= "%s"
	access_level 	= "developer"
	ldap_provider   = "%s"

}`, rInt, rInt, testLdapLink.CN, testLdapLink.Provider)
}

func testAccGitlabGroupLdapLinkUpdateConfig(rInt int, testLdapLink *gitlab.LDAPGroupLink) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
    name = "foo%d"
	path = "foo%d"
	description = "Terraform acceptance test - Group LDAP Links 2"
}

resource "gitlab_group_ldap_link" "foo" {
    group_id 		= "${gitlab_group.foo.id}"
    cn				= "%s"
	access_level 	= "maintainer"
	ldap_provider   = "%s"
}`, rInt, rInt, testLdapLink.CN, testLdapLink.Provider)
}

func testAccGitlabGroupLdapLinkForceCreateConfig(rInt int, testLdapLink *gitlab.LDAPGroupLink) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
    name = "foo%d"
	path = "foo%d"
	description = "Terraform acceptance test - Group LDAP Links 3"
}

resource "gitlab_group_ldap_link" "foo" {
    group_id 		= "${gitlab_group.foo.id}"
    cn				= "%s"
	access_level 	= "maintainer"
	ldap_provider   = "%s"
}

resource "gitlab_group_ldap_link" "bar" {
    group_id 		= "${gitlab_group.foo.id}"
    cn				= "%s"
	access_level 	= "developer"
	ldap_provider   = "%s"
	force			= true
}`, rInt, rInt, testLdapLink.CN, testLdapLink.Provider, testLdapLink.CN, testLdapLink.Provider)
}
