package gitlab

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabGroupLdapLink() *schema.Resource {
	acceptedAccessLevels := make([]string, 0, len(accessLevelID))
	for k := range accessLevelID {
		acceptedAccessLevels = append(acceptedAccessLevels, k)
	}
	// lintignore: XR002 // TODO: Resolve this tfproviderlint issue
	return &schema.Resource{
		Create: resourceGitlabGroupLdapLinkCreate,
		Read:   resourceGitlabGroupLdapLinkRead,
		Delete: resourceGitlabGroupLdapLinkDelete,

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Using the friendlier "access_level" here instead of the GitLab API "group_access".
			"access_level": {
				Type:         schema.TypeString,
				ValidateFunc: validateValueFunc(acceptedAccessLevels),
				Required:     true,
				ForceNew:     true,
			},
			// Changing GitLab API parameter "provider" to "ldap_provider" to avoid clashing with the Terraform "provider" key word
			"ldap_provider": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"force": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceGitlabGroupLdapLinkCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	groupId := d.Get("group_id").(string)
	cn := d.Get("cn").(string)
	group_access := gitlab.AccessLevelValue(accessLevelNameToValue[d.Get("access_level").(string)])
	ldap_provider := d.Get("ldap_provider").(string)
	force := d.Get("force").(bool)

	options := &gitlab.AddGroupLDAPLinkOptions{
		CN:          &cn,
		GroupAccess: gitlab.AccessLevel(gitlab.AccessLevelValue(group_access)),
		Provider:    &ldap_provider,
	}

	if force {
		resourceGitlabGroupLdapLinkDelete(d, meta) // nolint // TODO: Resolve this golangci-lint issue: Error return value is not checked (errcheck)
	}

	log.Printf("[DEBUG] Create GitLab group LdapLink %s", d.Id())
	LdapLink, _, err := client.Groups.AddGroupLDAPLink(groupId, options)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(&LdapLink.Provider, &LdapLink.CN))

	return resourceGitlabGroupLdapLinkRead(d, meta)
}

func resourceGitlabGroupLdapLinkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	groupId := d.Get("group_id").(string)

	// Try to fetch all group links from GitLab
	log.Printf("[DEBUG] Read GitLab group LdapLinks %s", groupId)
	ldapLinks, _, err := client.Groups.ListGroupLDAPLinks(groupId, nil)
	if err != nil {
		// The read/GET API wasn't implemented in GitLab until version 12.8 (March 2020, well after the add and delete APIs).
		// If we 404, assume GitLab is at an older version and take things on faith.
		switch err.(type) { // nolint // TODO: Resolve this golangci-lint issue: S1034: assigning the result of this type assertion to a variable (switch err := err.(type)) could eliminate type assertions in switch cases (gosimple)
		case *gitlab.ErrorResponse:
			if err.(*gitlab.ErrorResponse).Response.StatusCode == 404 { // nolint // TODO: Resolve this golangci-lint issue: S1034(related information): could eliminate this type assertion (gosimple)
				log.Printf("[WARNING] This GitLab instance doesn't have the GET API for group_ldap_sync.  Please upgrade to 12.8 or later for best results.")
			} else {
				return err
			}
		default:
			return err
		}
	}

	// If we got here and don't have links, assume GitLab is below version 12.8 and skip the check
	if ldapLinks != nil {
		// Check if the LDAP link exists in the returned list of links
		found := false
		for _, ldapLink := range ldapLinks {
			if buildTwoPartID(&ldapLink.Provider, &ldapLink.CN) == d.Id() {
				d.Set("group_id", groupId)
				d.Set("cn", ldapLink.CN)
				d.Set("group_access", ldapLink.GroupAccess)
				d.Set("ldap_provider", ldapLink.Provider)
				found = true
				break
			}
		}

		if !found {
			d.SetId("")
			return errors.New(fmt.Sprintf("LdapLink %s does not exist.", d.Id())) // nolint // TODO: Resolve this golangci-lint issue: S1028: should use fmt.Errorf(...) instead of errors.New(fmt.Sprintf(...)) (gosimple)
		}
	}

	return nil
}

func resourceGitlabGroupLdapLinkDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	groupId := d.Get("group_id").(string)
	cn := d.Get("cn").(string)
	ldap_provider := d.Get("ldap_provider").(string)

	log.Printf("[DEBUG] Delete GitLab group LdapLink %s", d.Id())
	_, err := client.Groups.DeleteGroupLDAPLinkForProvider(groupId, ldap_provider, cn)
	if err != nil {
		switch err.(type) { // nolint // TODO: Resolve this golangci-lint issue: S1034: assigning the result of this type assertion to a variable (switch err := err.(type)) could eliminate type assertions in switch cases (gosimple)
		case *gitlab.ErrorResponse:
			// Ignore LDAP links that don't exist
			if strings.Contains(string(err.(*gitlab.ErrorResponse).Message), "Linked LDAP group not found") { // nolint // TODO: Resolve this golangci-lint issue: S1034(related information): could eliminate this type assertion (gosimple)
				log.Printf("[WARNING] %s", err)
			} else {
				return err
			}
		default:
			return err
		}
	}

	return nil
}
