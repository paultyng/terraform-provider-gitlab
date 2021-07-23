package gitlab

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/httpclient"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {

	// The actual provider
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITLAB_TOKEN", nil),
				Description: descriptions["token"],
			},
			"base_url": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("GITLAB_BASE_URL", ""),
				Description:  descriptions["base_url"],
				ValidateFunc: validateApiURLVersion,
			},
			"cacert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["cacert_file"],
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["insecure"],
			},
			"client_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["client_cert"],
			},
			"client_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["client_key"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"gitlab_group":            dataSourceGitlabGroup(),
			"gitlab_group_membership": dataSourceGitlabGroupMembership(),
			"gitlab_project":          dataSourceGitlabProject(),
			"gitlab_projects":         dataSourceGitlabProjects(),
			"gitlab_user":             dataSourceGitlabUser(),
			"gitlab_users":            dataSourceGitlabUsers(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"gitlab_branch_protection":          resourceGitlabBranchProtection(),
			"gitlab_tag_protection":             resourceGitlabTagProtection(),
			"gitlab_group":                      resourceGitlabGroup(),
			"gitlab_project":                    resourceGitlabProject(),
			"gitlab_label":                      resourceGitlabLabel(),
			"gitlab_group_label":                resourceGitlabGroupLabel(),
			"gitlab_pipeline_schedule":          resourceGitlabPipelineSchedule(),
			"gitlab_pipeline_schedule_variable": resourceGitlabPipelineScheduleVariable(),
			"gitlab_pipeline_trigger":           resourceGitlabPipelineTrigger(),
			"gitlab_project_hook":               resourceGitlabProjectHook(),
			"gitlab_deploy_key":                 resourceGitlabDeployKey(),
			"gitlab_deploy_key_enable":          resourceGitlabDeployEnableKey(),
			"gitlab_deploy_token":               resourceGitlabDeployToken(),
			"gitlab_user":                       resourceGitlabUser(),
			"gitlab_project_membership":         resourceGitlabProjectMembership(),
			"gitlab_group_membership":           resourceGitlabGroupMembership(),
			"gitlab_project_variable":           resourceGitlabProjectVariable(),
			"gitlab_group_variable":             resourceGitlabGroupVariable(),
			"gitlab_project_cluster":            resourceGitlabProjectCluster(),
			"gitlab_service_slack":              resourceGitlabServiceSlack(),
			"gitlab_service_jira":               resourceGitlabServiceJira(),
			"gitlab_service_github":             resourceGitlabServiceGithub(),
			"gitlab_service_pipelines_email":    resourceGitlabServicePipelinesEmail(),
			"gitlab_project_share_group":        resourceGitlabProjectShareGroup(),
			"gitlab_group_cluster":              resourceGitlabGroupCluster(),
			"gitlab_group_ldap_link":            resourceGitlabGroupLdapLink(),
			"gitlab_instance_cluster":           resourceGitlabInstanceCluster(),
			"gitlab_project_mirror":             resourceGitlabProjectMirror(),
			"gitlab_project_level_mr_approvals": resourceGitlabProjectLevelMRApprovals(),
			"gitlab_project_approval_rule":      resourceGitlabProjectApprovalRule(),
			"gitlab_instance_variable":          resourceGitlabInstanceVariable(),
			"gitlab_project_freeze_period":      resourceGitlabProjectFreezePeriod(),
			"gitlab_group_share_group":          resourceGitlabGroupShareGroup(),
			"gitlab_project_badge":              resourceGitlabProjectBadge(),
			"gitlab_group_badge":                resourceGitlabGroupBadge(),
		},
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		return providerConfigure(provider, d)
	}

	return provider
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"token": "The OAuth token used to connect to GitLab.",

		"base_url": "The GitLab Base API URL",

		"cacert_file": "A file containing the ca certificate to use in case ssl certificate is not from a standard chain",

		"insecure": "Disable SSL verification of API calls",

		"client_cert": "File path to client certificate when GitLab instance is behind company proxy. File  must contain PEM encoded data.",

		"client_key": "File path to client key when GitLab instance is behind company proxy. File must contain PEM encoded data.",
	}
}

func providerConfigure(p *schema.Provider, d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Token:      d.Get("token").(string),
		BaseURL:    d.Get("base_url").(string),
		CACertFile: d.Get("cacert_file").(string),
		Insecure:   d.Get("insecure").(bool),
		ClientCert: d.Get("client_cert").(string),
		ClientKey:  d.Get("client_key").(string),
	}

	client, err := config.Client()
	if err != nil {
		return nil, err
	}

	// NOTE: httpclient.TerraformUserAgent is deprecated and removed in Terraform SDK v2
	// After upgrading the SDK to v2 replace with p.UserAgent("terraform-provider-gitlab")
	client.UserAgent = httpclient.TerraformUserAgent(p.TerraformVersion) + " terraform-provider-gitlab" // nolint // TODO: Resolve this golangci-lint issue: SA1019: httpclient.TerraformUserAgent is deprecated: This will be removed in v2 without replacement. If you need its functionality, you can copy it or reference the v1 package. (staticcheck)

	return client, err
}

func validateApiURLVersion(value interface{}, key string) (ws []string, es []error) {
	v := value.(string)
	if strings.HasSuffix(v, "/api/v3") || strings.HasSuffix(v, "/api/v3/") {
		es = append(es, fmt.Errorf("terraform-provider-gitlab does not support v3 api; please upgrade to /api/v4 in %s", v))
	}
	return
}
