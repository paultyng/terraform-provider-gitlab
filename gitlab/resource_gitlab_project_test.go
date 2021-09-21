// lintignore: AT012 // TODO: Resolve this tfproviderlint issue

package gitlab

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	gitlab "github.com/xanzy/go-gitlab"
)

type testAccGitlabProjectExpectedAttributes struct {
	DefaultBranch string
}

func TestAccGitlabProject_basic(t *testing.T) {
	var received, defaults, defaultsMainBranch gitlab.Project
	rInt := acctest.RandInt()

	defaults = gitlab.Project{
		Namespace:                        &gitlab.ProjectNamespace{ID: 0},
		Name:                             fmt.Sprintf("foo-%d", rInt),
		Path:                             fmt.Sprintf("foo.%d", rInt),
		Description:                      "Terraform acceptance tests",
		TagList:                          []string{"tag1"},
		RequestAccessEnabled:             true,
		IssuesEnabled:                    true,
		MergeRequestsEnabled:             true,
		JobsEnabled:                      true,
		ApprovalsBeforeMerge:             0,
		WikiEnabled:                      true,
		SnippetsEnabled:                  true,
		ContainerRegistryEnabled:         true,
		LFSEnabled:                       true,
		SharedRunnersEnabled:             true,
		Visibility:                       gitlab.PublicVisibility,
		MergeMethod:                      gitlab.FastForwardMerge,
		OnlyAllowMergeIfPipelineSucceeds: true,
		OnlyAllowMergeIfAllDiscussionsAreResolved: true,
		AllowMergeOnSkippedPipeline:               false,
		Archived:                                  false, // needless, but let's make this explicit
		PackagesEnabled:                           true,
		PagesAccessLevel:                          gitlab.PublicAccessControl,
		BuildCoverageRegex:                        "foo",
		CIConfigPath:                              ".gitlab-ci.yml@mynamespace/myproject",
	}

	defaultsMainBranch = defaults
	defaultsMainBranch.DefaultBranch = "main"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectDestroy,
		Steps: []resource.TestStep{
			// Create a project with all the features on (note: "archived" is "false")
			{
				Config: testAccGitlabProjectConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.foo", &received),
					testAccCheckAggregateGitlabProject(&defaults, &received),
				),
			},
			// Update the project to turn the features off (note: "archived" is "true")
			{
				Config: testAccGitlabProjectUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.foo", &received),
					testAccCheckAggregateGitlabProject(&gitlab.Project{
						Namespace:                        &gitlab.ProjectNamespace{ID: 0},
						Name:                             fmt.Sprintf("foo-%d", rInt),
						Path:                             fmt.Sprintf("foo.%d", rInt),
						Description:                      "Terraform acceptance tests!",
						TagList:                          []string{"tag1", "tag2"},
						JobsEnabled:                      false,
						ApprovalsBeforeMerge:             0,
						RequestAccessEnabled:             false,
						ContainerRegistryEnabled:         false,
						LFSEnabled:                       false,
						SharedRunnersEnabled:             false,
						Visibility:                       gitlab.PublicVisibility,
						MergeMethod:                      gitlab.FastForwardMerge,
						OnlyAllowMergeIfPipelineSucceeds: true,
						OnlyAllowMergeIfAllDiscussionsAreResolved: true,
						AllowMergeOnSkippedPipeline:               false,
						Archived:                                  true,
						PackagesEnabled:                           false,
						PagesAccessLevel:                          gitlab.DisabledAccessControl,
						BuildCoverageRegex:                        "bar",
					}, &received),
				),
			},
			// Update the project to turn the features on again (note: "archived" is "false")
			{
				Config: testAccGitlabProjectConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.foo", &received),
					testAccCheckAggregateGitlabProject(&defaults, &received),
				),
			},
			// Update the project creating the default branch
			{
				// Get the ID from the project data at the previous step
				SkipFunc: testAccGitlabProjectConfigDefaultBranchSkipFunc(&received, "main"),
				Config:   testAccGitlabProjectConfigDefaultBranch(rInt, "main"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.foo", &received),
					testAccCheckAggregateGitlabProject(&defaultsMainBranch, &received),
				),
			},
			// Test import without push rules (checks read function)
			{
				ResourceName:      "gitlab_project.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Add all push rules to an existing project
			{
				SkipFunc: isRunningInCE,
				Config: testAccGitlabProjectConfigPushRules(rInt, `
author_email_regex = "foo_author"
branch_name_regex = "foo_branch"
commit_message_regex = "foo_commit"
commit_message_negative_regex = "foo_not_commit"
file_name_regex = "foo_file"
commit_committer_check = true
deny_delete_tag = true
member_check = true
prevent_secrets = true
reject_unsigned_commits = true
max_file_size = 123
`),
				Check: testAccCheckGitlabProjectPushRules("gitlab_project.foo", &gitlab.ProjectPushRules{
					AuthorEmailRegex:           "foo_author",
					BranchNameRegex:            "foo_branch",
					CommitMessageRegex:         "foo_commit",
					CommitMessageNegativeRegex: "foo_not_commit",
					FileNameRegex:              "foo_file",
					CommitCommitterCheck:       true,
					DenyDeleteTag:              true,
					MemberCheck:                true,
					PreventSecrets:             true,
					RejectUnsignedCommits:      true,
					MaxFileSize:                123,
				}),
			},
			// Test import with a all push rules defined (checks read function)
			{
				SkipFunc:          isRunningInCE,
				ResourceName:      "gitlab_project.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update some push rules but not others
			{
				SkipFunc: isRunningInCE,
				Config: testAccGitlabProjectConfigPushRules(rInt, `
author_email_regex = "foo_author"
branch_name_regex = "foo_branch"
commit_message_regex = "foo_commit"
commit_message_negative_regex = "foo_not_commit"
file_name_regex = "foo_file_2"
commit_committer_check = true
deny_delete_tag = true
member_check = false
prevent_secrets = true
reject_unsigned_commits = true
max_file_size = 1234
`),
				Check: testAccCheckGitlabProjectPushRules("gitlab_project.foo", &gitlab.ProjectPushRules{
					AuthorEmailRegex:           "foo_author",
					BranchNameRegex:            "foo_branch",
					CommitMessageRegex:         "foo_commit",
					CommitMessageNegativeRegex: "foo_not_commit",
					FileNameRegex:              "foo_file_2",
					CommitCommitterCheck:       true,
					DenyDeleteTag:              true,
					MemberCheck:                false,
					PreventSecrets:             true,
					RejectUnsignedCommits:      true,
					MaxFileSize:                1234,
				}),
			},
			// Try to add push rules to an existing project in CE
			{
				SkipFunc:    isRunningInEE,
				Config:      testAccGitlabProjectConfigPushRules(rInt, `author_email_regex = "foo_author"`),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("Project push rules are not supported in your version of GitLab")),
			},
			// Update push rules
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabProjectConfigPushRules(rInt, `author_email_regex = "foo_author"`),
				Check: testAccCheckGitlabProjectPushRules("gitlab_project.foo", &gitlab.ProjectPushRules{
					AuthorEmailRegex: "foo_author",
				}),
			},
			// Remove the push_rules block entirely.
			// NOTE: The push rules will still exist upstream because the push_rules block is computed.
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabProjectConfigDefaultBranch(rInt, "main"),
				Check: testAccCheckGitlabProjectPushRules("gitlab_project.foo", &gitlab.ProjectPushRules{
					AuthorEmailRegex: "foo_author",
				}),
			},
			// Add different push rules after the block was removed previously
			{
				SkipFunc: isRunningInCE,
				Config:   testAccGitlabProjectConfigPushRules(rInt, `branch_name_regex = "(feature|hotfix)\\/*"`),
				Check: testAccCheckGitlabProjectPushRules("gitlab_project.foo", &gitlab.ProjectPushRules{
					BranchNameRegex: `(feature|hotfix)\/*`,
				}),
			},
			// Destroy the project so we can next test creating a project with push rules simultaneously
			{
				Config:  testAccGitlabProjectConfigDefaultBranch(rInt, "main"),
				Destroy: true,
				Check:   testAccCheckGitlabProjectDestroy,
			},
			// Create a new project with push rules
			{
				SkipFunc: isRunningInCE,
				Config: testAccGitlabProjectConfigPushRules(rInt, `
author_email_regex = "foo_author"
max_file_size = 123
`),
				Check: testAccCheckGitlabProjectPushRules("gitlab_project.foo", &gitlab.ProjectPushRules{
					AuthorEmailRegex: "foo_author",
					MaxFileSize:      123,
				}),
			},
			// Try to create a new project with all push rules in CE
			{
				SkipFunc:    isRunningInEE,
				Config:      testAccGitlabProjectConfigPushRules(rInt, `author_email_regex = "foo_author"`),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("Project push rules are not supported in your version of GitLab")),
			},
			// Create a project using template name
			{
				Config: testAccGitlabProjectConfigTemplateName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.template-name", &received),
					testAccCheckGitlabProjectDefaultBranch(&received, &testAccGitlabProjectExpectedAttributes{
						DefaultBranch: "master",
					}),
					func(state *terraform.State) error {
						client := testAccProvider.Meta().(*gitlab.Client)

						projectID := state.RootModule().Resources["gitlab_project.template-name"].Primary.ID

						_, _, err := client.RepositoryFiles.GetFile(projectID, ".ruby-version", &gitlab.GetFileOptions{Ref: gitlab.String("master")}, nil)
						if err != nil {
							return fmt.Errorf("failed to get '.ruby-version' file from template project: %w", err)
						}

						return nil
					},
				),
			},
			// Create a project using custom template name
			{
				Config:   testAccGitlabProjectConfigTemplateNameCustom(rInt),
				SkipFunc: isRunningInCE,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.template-name-custom", &received),
					testAccCheckGitlabProjectDefaultBranch(&received, &testAccGitlabProjectExpectedAttributes{
						DefaultBranch: "master",
					}),
					func(state *terraform.State) error {
						client := testAccProvider.Meta().(*gitlab.Client)

						projectID := state.RootModule().Resources["gitlab_project.template-name-custom"].Primary.ID

						_, _, err := client.RepositoryFiles.GetFile(projectID, "Gemfile", &gitlab.GetFileOptions{Ref: gitlab.String("master")}, nil)
						if err != nil {
							return fmt.Errorf("failed to get 'Gemfile' file from template project: %w", err)
						}

						return nil
					},
				),
			},
			// Create a project using custom template project id
			{
				Config:   testAccGitlabProjectConfigTemplateProjectID(rInt),
				SkipFunc: isRunningInCE,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.template-id", &received),
					testAccCheckGitlabProjectDefaultBranch(&received, &testAccGitlabProjectExpectedAttributes{
						DefaultBranch: "master",
					}),
					func(state *terraform.State) error {
						client := testAccProvider.Meta().(*gitlab.Client)

						projectID := state.RootModule().Resources["gitlab_project.template-id"].Primary.ID

						_, _, err := client.RepositoryFiles.GetFile(projectID, "Rakefile", &gitlab.GetFileOptions{Ref: gitlab.String("master")}, nil)
						if err != nil {
							return fmt.Errorf("failed to get 'Rakefile' file from template project: %w", err)
						}

						return nil
					},
				),
			},
			// Update to original project config
			{
				Config: testAccGitlabProjectConfig(rInt),
			},
		},
	})
}

func TestAccGitlabProject_initializeWithReadme(t *testing.T) {
	var project gitlab.Project
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabProjectConfigInitializeWithReadme(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.foo", &project),
					testAccCheckGitlabProjectDefaultBranch(&project, nil),
				),
			},
		},
	})
}

func TestAccGitlabProject_willError(t *testing.T) {
	var received, defaults gitlab.Project
	rInt := acctest.RandInt()
	defaults = gitlab.Project{
		Namespace:                        &gitlab.ProjectNamespace{ID: 0},
		Name:                             fmt.Sprintf("foo-%d", rInt),
		Path:                             fmt.Sprintf("foo.%d", rInt),
		Description:                      "Terraform acceptance tests",
		TagList:                          []string{"tag1"},
		RequestAccessEnabled:             true,
		IssuesEnabled:                    true,
		MergeRequestsEnabled:             true,
		JobsEnabled:                      true,
		ApprovalsBeforeMerge:             0,
		WikiEnabled:                      true,
		SnippetsEnabled:                  true,
		ContainerRegistryEnabled:         true,
		LFSEnabled:                       true,
		SharedRunnersEnabled:             true,
		Visibility:                       gitlab.PublicVisibility,
		MergeMethod:                      gitlab.FastForwardMerge,
		OnlyAllowMergeIfPipelineSucceeds: true,
		OnlyAllowMergeIfAllDiscussionsAreResolved: true,
		PackagesEnabled:    true,
		PagesAccessLevel:   gitlab.PublicAccessControl,
		BuildCoverageRegex: "foo",
		CIConfigPath:       ".gitlab-ci.yml@mynamespace/myproject",
	}
	willError := defaults
	willError.TagList = []string{"notatag"}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectDestroy,
		Steps: []resource.TestStep{
			// Step0 Create a project
			{
				Config: testAccGitlabProjectConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.foo", &received),
					testAccCheckAggregateGitlabProject(&defaults, &received),
				),
			},
			// Step1 Verify that passing bad values will fail.
			{
				Config:      testAccGitlabProjectConfig(rInt),
				ExpectError: regexp.MustCompile(`\stags\sexpected\s.+notatag.+\sreceived`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAggregateGitlabProject(&willError, &received),
				),
			},
			// Step2 Reset
			{
				Config: testAccGitlabProjectConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.foo", &received),
					testAccCheckAggregateGitlabProject(&defaults, &received),
				),
			},
		},
	})
}

// lintignore: AT002 // TODO: Resolve this tfproviderlint issue
func TestAccGitlabProject_import(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabProjectConfig(rInt),
			},
			{
				ResourceName:      "gitlab_project.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGitlabProject_nestedImport(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabProjectInGroupConfig(rInt),
			},
			{
				ResourceName:      "gitlab_project.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGitlabProject_transfer(t *testing.T) {
	var transferred, received gitlab.Project
	rInt := acctest.RandInt()

	transferred = gitlab.Project{
		Namespace:                        &gitlab.ProjectNamespace{Name: fmt.Sprintf("foo2group-%d", rInt)},
		Name:                             fmt.Sprintf("foo-%d", rInt),
		Path:                             fmt.Sprintf("foo-%d", rInt),
		Description:                      "Terraform acceptance tests",
		TagList:                          []string{},
		RequestAccessEnabled:             true,
		IssuesEnabled:                    true,
		MergeRequestsEnabled:             true,
		JobsEnabled:                      true,
		ApprovalsBeforeMerge:             0,
		WikiEnabled:                      true,
		SnippetsEnabled:                  true,
		ContainerRegistryEnabled:         true,
		LFSEnabled:                       true,
		SharedRunnersEnabled:             true,
		Visibility:                       gitlab.PublicVisibility,
		MergeMethod:                      gitlab.NoFastForwardMerge,
		OnlyAllowMergeIfPipelineSucceeds: false,
		OnlyAllowMergeIfAllDiscussionsAreResolved: false,
		PackagesEnabled:    true,
		PagesAccessLevel:   gitlab.PrivateAccessControl,
		BuildCoverageRegex: "foo",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectDestroy,
		Steps: []resource.TestStep{
			// Create a project in a group
			{
				Config: testAccGitlabProjectInGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.foo", &received),
				),
			},
			// Create a second group and set the transfer the project to this group
			{
				Config: testAccGitlabProjectTransferBetweenGroups(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.foo", &received),
					testAccCheckAggregateGitlabProject(&transferred, &received),
				),
			},
		},
	})
}

// lintignore: AT002 // not a Terraform import test
func TestAccGitlabProject_importURL(t *testing.T) {
	// Since we do some manual setup in this test, we need to handle the test skip first.
	if os.Getenv(resource.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf("Acceptance tests skipped unless env '%s' set", resource.TestEnvVar))
	}

	client := testAccProvider.Meta().(*gitlab.Client)
	rInt := acctest.RandInt()

	// Create a base project for importing.
	baseProject, _, err := client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:       gitlab.String(fmt.Sprintf("base-%d", rInt)),
		Visibility: gitlab.Visibility(gitlab.PublicVisibility),
	})
	if err != nil {
		t.Fatalf("failed to create base project: %v", err)
	}

	defer client.Projects.DeleteProject(baseProject.ID) // nolint // TODO: Resolve this golangci-lint issue: Error return value of `client.Projects.DeleteProject` is not checked (errcheck)

	// Add a file to the base project, for later verifying the import.
	_, _, err = client.RepositoryFiles.CreateFile(baseProject.ID, "foo.txt", &gitlab.CreateFileOptions{
		Branch:        gitlab.String("main"),
		CommitMessage: gitlab.String("add file"),
		Content:       gitlab.String(""),
	})
	if err != nil {
		t.Fatalf("failed to commit file to base project: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabProjectConfigImportURL(rInt, baseProject.HTTPURLToRepo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("gitlab_project.imported", "import_url", baseProject.HTTPURLToRepo),
					func(state *terraform.State) error {
						projectID := state.RootModule().Resources["gitlab_project.imported"].Primary.ID

						_, _, err := client.RepositoryFiles.GetFile(projectID, "foo.txt", &gitlab.GetFileOptions{Ref: gitlab.String("main")}, nil)
						if err != nil {
							return fmt.Errorf("failed to get file from imported project: %w", err)
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccGitlabProject_initializeWithReadmeAndCustomDefaultBranch(t *testing.T) {
	var project gitlab.Project
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"

  initialize_with_readme = true
  default_branch         = "foo"
}`, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("gitlab_project.foo", "initialize_with_readme", "true"),
					resource.TestCheckResourceAttr("gitlab_project.foo", "default_branch", "foo"),
					testAccCheckGitlabProjectExists("gitlab_project.foo", &project),
					testAccCheckGitlabProjectDefaultBranch(&project, &testAccGitlabProjectExpectedAttributes{
						DefaultBranch: "foo",
					}),
				),
			},
		},
	})
}

type testAccGitlabProjectMirroredExpectedAttributes struct {
	Mirror                           bool
	MirrorTriggerBuilds              bool
	MirrorOverwritesDivergedBranches bool
	OnlyMirrorProtectedBranches      bool
}

func testAccCheckGitlabProjectMirroredAttributes(project *gitlab.Project, want *testAccGitlabProjectMirroredExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if project.Mirror != want.Mirror {
			return fmt.Errorf("got mirror %t; want %t", project.Mirror, want.Mirror)
		}

		if project.MirrorTriggerBuilds != want.MirrorTriggerBuilds {
			return fmt.Errorf("got mirror_trigger_builds %t; want %t", project.MirrorTriggerBuilds, want.MirrorTriggerBuilds)
		}

		if project.MirrorOverwritesDivergedBranches != want.MirrorOverwritesDivergedBranches {
			return fmt.Errorf("got mirror_overwrites_diverged_branches %t; want %t", project.MirrorOverwritesDivergedBranches, want.MirrorOverwritesDivergedBranches)
		}

		if project.OnlyMirrorProtectedBranches != want.OnlyMirrorProtectedBranches {
			return fmt.Errorf("got only_mirror_protected_branches %t; want %t", project.OnlyMirrorProtectedBranches, want.OnlyMirrorProtectedBranches)
		}
		return nil
	}
}

// lintignore: AT002 // not a Terraform import test
func TestAccGitlabProject_importURLMirrored(t *testing.T) {
	// Since we do some manual setup in this test, we need to handle the test skip first.
	if os.Getenv(resource.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf("Acceptance tests skipped unless env '%s' set", resource.TestEnvVar))
	}

	client := testAccProvider.Meta().(*gitlab.Client)
	var mirror gitlab.Project
	rInt := acctest.RandInt()

	// Create a base project for importing.
	baseProject, _, err := client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:       gitlab.String(fmt.Sprintf("base-%d", rInt)),
		Visibility: gitlab.Visibility(gitlab.PublicVisibility),
	})
	if err != nil {
		t.Fatalf("failed to create base project: %v", err)
	}

	defer client.Projects.DeleteProject(baseProject.ID) // nolint // TODO: Resolve this golangci-lint issue: Error return value of `client.Projects.DeleteProject` is not checked (errcheck)

	// Add a file to the base project, for later verifying the import.
	_, _, err = client.RepositoryFiles.CreateFile(baseProject.ID, "foo.txt", &gitlab.CreateFileOptions{
		Branch:        gitlab.String("main"),
		CommitMessage: gitlab.String("add file"),
		Content:       gitlab.String(""),
	})
	if err != nil {
		t.Fatalf("failed to commit file to base project: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGitlabProjectDestroy,
		Steps: []resource.TestStep{
			{
				// First, import, as mirrored
				Config:   testAccGitlabProjectConfigImportURLMirror(rInt, baseProject.HTTPURLToRepo),
				SkipFunc: isRunningInCE,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.imported", &mirror),
					resource.TestCheckResourceAttr("gitlab_project.imported", "import_url", baseProject.HTTPURLToRepo),
					testAccCheckGitlabProjectMirroredAttributes(&mirror, &testAccGitlabProjectMirroredExpectedAttributes{
						Mirror:                           true,
						MirrorTriggerBuilds:              true,
						MirrorOverwritesDivergedBranches: true,
						OnlyMirrorProtectedBranches:      true,
					}),

					func(state *terraform.State) error {
						projectID := state.RootModule().Resources["gitlab_project.imported"].Primary.ID

						_, _, err := client.RepositoryFiles.GetFile(projectID, "foo.txt", &gitlab.GetFileOptions{Ref: gitlab.String("main")}, nil)
						if err != nil {
							return fmt.Errorf("failed to get file from imported project: %w", err)
						}

						return nil
					},
				),
			},
			{
				// Second, disable all optional mirroring options
				Config:   testAccGitlabProjectConfigImportURLMirrorDisabledOptionals(rInt, baseProject.HTTPURLToRepo),
				SkipFunc: isRunningInCE,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.imported", &mirror),
					resource.TestCheckResourceAttr("gitlab_project.imported", "import_url", baseProject.HTTPURLToRepo),
					testAccCheckGitlabProjectMirroredAttributes(&mirror, &testAccGitlabProjectMirroredExpectedAttributes{
						Mirror:                           true,
						MirrorTriggerBuilds:              false,
						MirrorOverwritesDivergedBranches: false,
						OnlyMirrorProtectedBranches:      false,
					}),

					// Ensure the test file still is as expected
					func(state *terraform.State) error {
						projectID := state.RootModule().Resources["gitlab_project.imported"].Primary.ID

						_, _, err := client.RepositoryFiles.GetFile(projectID, "foo.txt", &gitlab.GetFileOptions{Ref: gitlab.String("main")}, nil)
						if err != nil {
							return fmt.Errorf("failed to get file from imported project: %w", err)
						}

						return nil
					},
				),
			},
			{
				// Third, disable mirroring, using the original ImportURL acceptance test
				Config:   testAccGitlabProjectConfigImportURLMirrorDisabled(rInt, baseProject.HTTPURLToRepo),
				SkipFunc: isRunningInCE,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectExists("gitlab_project.imported", &mirror),
					resource.TestCheckResourceAttr("gitlab_project.imported", "import_url", baseProject.HTTPURLToRepo),
					testAccCheckGitlabProjectMirroredAttributes(&mirror, &testAccGitlabProjectMirroredExpectedAttributes{
						Mirror:                           false,
						MirrorTriggerBuilds:              false,
						MirrorOverwritesDivergedBranches: false,
						OnlyMirrorProtectedBranches:      false,
					}),

					// Ensure the test file still is as expected
					func(state *terraform.State) error {
						projectID := state.RootModule().Resources["gitlab_project.imported"].Primary.ID

						_, _, err := client.RepositoryFiles.GetFile(projectID, "foo.txt", &gitlab.GetFileOptions{Ref: gitlab.String("main")}, nil)
						if err != nil {
							return fmt.Errorf("failed to get file from imported project: %w", err)
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccGitlabProjec_templateMutualExclusiveNameAndID(t *testing.T) {
	rInt := acctest.RandInt()

	// lintignore: AT001 // TODO: Resolve this tfproviderlint issue
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckMutualExclusiveNameAndID(rInt),
				SkipFunc:    isRunningInCE,
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`"template_project_id": conflicts with template_name`)),
			},
		},
	})
}

func testAccCheckGitlabProjectExists(n string, project *gitlab.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}
		repoName := rs.Primary.ID
		if repoName == "" {
			return fmt.Errorf("No project ID is set")
		}
		conn := testAccProvider.Meta().(*gitlab.Client)
		if g, _, err := conn.Projects.GetProject(repoName, nil); err == nil {
			*project = *g
		}
		return err
	}
}

func testAccCheckGitlabProjectDestroy(s *terraform.State) error {
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

func testAccCheckAggregateGitlabProject(expected, received *gitlab.Project) resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	testResource := resourceGitlabProject()
	expectedData := testResource.TestResourceData()
	receivedData := testResource.TestResourceData()
	for a, v := range testResource.Schema {
		attribute := a
		attrValue := v
		checks = append(checks, func(_ *terraform.State) error {
			if attrValue.Computed {
				if attrDefault, err := attrValue.DefaultValue(); err == nil {
					if attrDefault == nil {
						return nil // Skipping because we have no way of pre-computing computed vars
					}
				} else {
					return err
				}
			}

			if err := resourceGitlabProjectSetToState(expectedData, expected); err != nil {
				return err
			}

			if err := resourceGitlabProjectSetToState(receivedData, received); err != nil {
				return err
			}

			return testAccCompareGitLabAttribute(attribute, expectedData, receivedData)
		})
	}
	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func testAccCheckGitlabProjectDefaultBranch(project *gitlab.Project, want *testAccGitlabProjectExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if want != nil && project.DefaultBranch != want.DefaultBranch {
			return fmt.Errorf("got default branch %q; want %q", project.DefaultBranch, want.DefaultBranch)
		}

		client := testAccProvider.Meta().(*gitlab.Client)

		branches, _, err := client.Branches.ListBranches(project.ID, nil)
		if err != nil {
			return fmt.Errorf("failed to list branches: %w", err)
		}

		if len(branches) != 1 {
			return fmt.Errorf("expected 1 branch for new project; found %d", len(branches))
		}

		if !branches[0].Protected {
			return errors.New("expected default branch to be protected")
		}

		return nil
	}
}

func testAccCheckGitlabProjectPushRules(name string, wantPushRules *gitlab.ProjectPushRules) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccProvider.Meta().(*gitlab.Client)
		projectResource := state.RootModule().Resources[name].Primary

		gotPushRules, _, err := client.Projects.GetProjectPushRules(projectResource.ID, nil)
		if err != nil {
			return err
		}

		var messages []string

		if gotPushRules.AuthorEmailRegex != wantPushRules.AuthorEmailRegex {
			messages = append(messages, fmt.Sprintf("author_email_regex (got: %q, wanted: %q)",
				gotPushRules.AuthorEmailRegex, wantPushRules.AuthorEmailRegex))
		}

		if gotPushRules.BranchNameRegex != wantPushRules.BranchNameRegex {
			messages = append(messages, fmt.Sprintf("branch_name_regex (got: %q, wanted: %q)",
				gotPushRules.BranchNameRegex, wantPushRules.BranchNameRegex))
		}

		if gotPushRules.CommitMessageRegex != wantPushRules.CommitMessageRegex {
			messages = append(messages, fmt.Sprintf("commit_message_regex (got: %q, wanted: %q)",
				gotPushRules.CommitMessageRegex, wantPushRules.CommitMessageRegex))
		}

		if gotPushRules.CommitMessageNegativeRegex != wantPushRules.CommitMessageNegativeRegex {
			messages = append(messages, fmt.Sprintf("commit_message_negative_regex (got: %q, wanted: %q)",
				gotPushRules.CommitMessageNegativeRegex, wantPushRules.CommitMessageNegativeRegex))
		}

		if gotPushRules.FileNameRegex != wantPushRules.FileNameRegex {
			messages = append(messages, fmt.Sprintf("file_name_regex (got: %q, wanted: %q)",
				gotPushRules.FileNameRegex, wantPushRules.FileNameRegex))
		}

		if gotPushRules.CommitCommitterCheck != wantPushRules.CommitCommitterCheck {
			messages = append(messages, fmt.Sprintf("commit_committer_check (got: %t, wanted: %t)",
				gotPushRules.CommitCommitterCheck, wantPushRules.CommitCommitterCheck))
		}

		if gotPushRules.DenyDeleteTag != wantPushRules.DenyDeleteTag {
			messages = append(messages, fmt.Sprintf("deny_delete_tag (got: %t, wanted: %t)",
				gotPushRules.DenyDeleteTag, wantPushRules.DenyDeleteTag))
		}

		if gotPushRules.MemberCheck != wantPushRules.MemberCheck {
			messages = append(messages, fmt.Sprintf("member_check (got: %t, wanted: %t)",
				gotPushRules.MemberCheck, wantPushRules.MemberCheck))
		}

		if gotPushRules.PreventSecrets != wantPushRules.PreventSecrets {
			messages = append(messages, fmt.Sprintf("prevent_secrets (got: %t, wanted: %t)",
				gotPushRules.PreventSecrets, wantPushRules.PreventSecrets))
		}

		if gotPushRules.RejectUnsignedCommits != wantPushRules.RejectUnsignedCommits {
			messages = append(messages, fmt.Sprintf("reject_unsigned_commits (got: %t, wanted: %t)",
				gotPushRules.RejectUnsignedCommits, wantPushRules.RejectUnsignedCommits))
		}

		if gotPushRules.MaxFileSize != wantPushRules.MaxFileSize {
			messages = append(messages, fmt.Sprintf("max_file_size (got: %d, wanted: %d)",
				gotPushRules.MaxFileSize, wantPushRules.MaxFileSize))
		}

		if len(messages) > 0 {
			return fmt.Errorf("unexpected push_rules:\n\t- %s", strings.Join(messages, "\n\t- "))
		}

		return nil
	}
}

func testAccGitlabProjectInGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name = "foogroup-%d"
  path = "foogroup-%d"
  visibility_level = "public"
}

resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"
  namespace_id = "${gitlab_group.foo.id}"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
  build_coverage_regex = "foo"
}
	`, rInt, rInt, rInt)
}

func testAccGitlabProjectTransferBetweenGroups(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name = "foogroup-%d"
  path = "foogroup-%d"
  visibility_level = "public"
}

resource "gitlab_group" "foo2" {
  name = "foo2group-%d"
  path = "foo2group-%d"
  visibility_level = "public"
}

resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"
  namespace_id = "${gitlab_group.foo2.id}"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
  build_coverage_regex = "foo"
}
	`, rInt, rInt, rInt, rInt, rInt)
}

func testAccGitlabProjectConfigDefaultBranch(rInt int, defaultBranch string) string {
	defaultBranchStatement := ""

	if len(defaultBranch) > 0 {
		defaultBranchStatement = fmt.Sprintf("default_branch = \"%s\"", defaultBranch)
	}

	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  path = "foo.%d"
  description = "Terraform acceptance tests"

  %s

  tags = [
	"tag1",
  ]

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
  merge_method = "ff"
  only_allow_merge_if_pipeline_succeeds = true
  only_allow_merge_if_all_discussions_are_resolved = true
  pages_access_level = "public"
  build_coverage_regex = "foo"
  ci_config_path = ".gitlab-ci.yml@mynamespace/myproject"
}
	`, rInt, rInt, defaultBranchStatement)
}

func testAccGitlabProjectConfigDefaultBranchSkipFunc(project *gitlab.Project, defaultBranch string) func() (bool, error) {
	return func() (bool, error) {
		conn := testAccProvider.Meta().(*gitlab.Client)

		// Commit data
		commitMessage := "Initial Commit"
		commitFile := "file.txt"
		commitFileAction := gitlab.FileCreate
		commitActions := []*gitlab.CommitActionOptions{
			{
				Action:   &commitFileAction,
				FilePath: &commitFile,
				Content:  &commitMessage,
			},
		}
		options := &gitlab.CreateCommitOptions{
			Branch:        &defaultBranch,
			CommitMessage: &commitMessage,
			Actions:       commitActions,
		}

		_, _, err := conn.Commits.CreateCommit(project.ID, options)

		return false, err
	}
}

func testAccGitlabProjectConfig(rInt int) string {
	return testAccGitlabProjectConfigDefaultBranch(rInt, "")
}

func testAccGitlabProjectUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  path = "foo.%d"
  description = "Terraform acceptance tests!"

  tags = [
	"tag1",
	"tag2",
  ]

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
  merge_method = "ff"
  only_allow_merge_if_pipeline_succeeds = true
  only_allow_merge_if_all_discussions_are_resolved = true
  allow_merge_on_skipped_pipeline = false

  request_access_enabled = false
  issues_enabled = false
  merge_requests_enabled = false
  pipelines_enabled = false
  approvals_before_merge = 0
  wiki_enabled = false
  snippets_enabled = false
  container_registry_enabled = false
  lfs_enabled = false
  shared_runners_enabled = false
  archived = true
  packages_enabled = false
  pages_access_level = "disabled"
  build_coverage_regex = "bar"
}
	`, rInt, rInt)
}

func testAccGitlabProjectConfigInitializeWithReadme(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  path = "foo.%d"
  description = "Terraform acceptance tests"
  initialize_with_readme = true
}
	`, rInt, rInt)
}

func testAccGitlabProjectConfigImportURL(rInt int, importURL string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "imported" {
  name = "imported-%d"
  default_branch = "main"
  import_url = "%s"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}
`, rInt, importURL)
}

func testAccGitlabProjectConfigImportURLMirror(rInt int, importURL string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "imported" {
  name = "imported-%d"
  default_branch = "main"
  import_url = "%s"
  mirror = true
  mirror_trigger_builds = true
  mirror_overwrites_diverged_branches = true
  only_mirror_protected_branches = true

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}
`, rInt, importURL)
}

func testAccGitlabProjectConfigImportURLMirrorDisabledOptionals(rInt int, importURL string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "imported" {
  name = "imported-%d"
  default_branch = "main"
  import_url = "%s"
  mirror = true
  mirror_trigger_builds = false
  mirror_overwrites_diverged_branches = false
  only_mirror_protected_branches = false

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}
`, rInt, importURL)
}

func testAccGitlabProjectConfigImportURLMirrorDisabled(rInt int, importURL string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "imported" {
  name = "imported-%d"
  default_branch = "main"
  import_url = "%s"
  mirror = false
  mirror_trigger_builds = false
  mirror_overwrites_diverged_branches = false
  only_mirror_protected_branches = false

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}
`, rInt, importURL)
}

func testAccGitlabProjectConfigPushRules(rInt int, pushRules string) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%[1]d"
  path = "foo.%[1]d"
  description = "Terraform acceptance tests"

  push_rules {
%[2]s
  }

  # So that acceptance tests can be run in a gitlab organization with no billing.
  visibility_level = "public"
}
	`, rInt, pushRules)
}

func testAccGitlabProjectConfigTemplateName(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "template-name" {
  name = "template-name-%d"
  path = "template-name.%d"
  description = "Terraform acceptance tests"
  template_name = "rails"
  default_branch = "master"
}
	`, rInt, rInt)
}

// 2020-09-07: Currently Gitlab (version 13.3.6 ) doesn't allow in admin API
// ability to set a group as instance level templates.
// To test resource_gitlab_project_test template features we add
// group, project myrails and admin settings directly in scripts/healthcheck-and-setup.sh
// Once Gitlab add admin template in API we could manage group/project/settings
// directly in tests like TestAccGitlabProject_basic.
func testAccGitlabProjectConfigTemplateNameCustom(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "template-name-custom" {
  name = "template-name-custom-%d"
  path = "template-name-custom.%d"
  description = "Terraform acceptance tests"
  template_name = "myrails"
  use_custom_template = true
  default_branch = "master"
}
	`, rInt, rInt)
}

func testAccGitlabProjectConfigTemplateProjectID(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "template-id" {
  name = "template-id-%d"
  path = "template-id.%d"
  description = "Terraform acceptance tests"
  template_project_id = 999
  use_custom_template = true
  default_branch = "master"
}
	`, rInt, rInt)
}

func testAccCheckMutualExclusiveNameAndID(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "template-mutual-exclusive" {
  name = "template-mutual-exclusive-%d"
  path = "template-mutual-exclusive.%d"
  description = "Terraform acceptance tests"
  template_name = "rails"
  template_project_id = 999
  use_custom_template = true
  default_branch = "master"
}
	`, rInt, rInt)
}
