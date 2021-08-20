package gitlab

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/xanzy/go-gitlab"
	"log"
	"strconv"
)

func resourceGitlabManagedLicense() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabManagedLicenseCreate,
		Read:   resourceGitlabManagedLicenseRead,
		Update: resourceGitlabManagedLicenseUpdate,
		Delete: resourceGitlabManagedLicenseDelete,

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				// GitLab's edit API doesn't allow us to edit the name, only
				// the approval status.
				ForceNew: true,
			},
			"approval_status": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				ValidateFunc: validation.StringInSlice([]string{"approved", "blacklisted"}, true),
			},
		},
	}
}

func resourceGitlabManagedLicenseCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)

	options := &gitlab.AddManagedLicenseOptions{
		Name:           gitlab.String(d.Get("name").(string)),
		ApprovalStatus: stringToApprovalStatus(d.Get("approval_status").(string)),
	}

	log.Printf("[DEBUG] create gitlab Managed License on Project %s, with Name %s", project, *options.Name)

	addManagedLicense, _, err := client.ManagedLicenses.AddManagedLicense(project, options)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(addManagedLicense.ID))
	return resourceGitlabManagedLicenseRead(d, meta)
}

func resourceGitlabManagedLicenseDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	licenseId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("%s cannot be converted to int", d.Id())
	}

	log.Printf("[DEBUG] Delete gitlab Managed License %s", d.Id())
	_, err = client.ManagedLicenses.DeleteManagedLicense(project, licenseId)
	if err != nil {
		return fmt.Errorf("failed to delete managed license %d for projec %s: %w", licenseId, project, err)
	}

	return nil
}

func resourceGitlabManagedLicenseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	id, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("%s cannot be converted to int", d.Id())
	}
	log.Printf("[DEBUG] read gitlab Managed License for project/id %s/%d", project, id)

	license, _, err := client.ManagedLicenses.GetManagedLicense(project, id)
	if err != nil {
		return fmt.Errorf("%s cannot be converted to int", d.Id())
	}

	d.Set("project", license.ID)
	d.Set("name", license.Name)
	d.Set("approval_status", license.ApprovalStatus)

	return nil
}

func resourceGitlabManagedLicenseUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	licenseId, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("%s cannot be converted to int", d.Id())
	}

	opts := &gitlab.EditManagedLicenceOptions{
		ApprovalStatus: stringToApprovalStatus(d.Get("approval_status").(string)),
	}

	if d.HasChange("approval_status") {
		opts.ApprovalStatus = stringToApprovalStatus(d.Get("approval_status").(string))
	}

	log.Printf("[DEBUG] update gitlab Managed License %s", d.Id())
	_, _, err = client.ManagedLicenses.EditManagedLicense(project, licenseId, opts)
	if err != nil {
		return err
	}

	return resourceGitlabManagedLicenseRead(d, meta)
}

// Convert the incoming string into the proper constant value for passing into the API.
func stringToApprovalStatus(s string) *gitlab.LicenseApprovalStatusValue {
	lookup := map[string]gitlab.LicenseApprovalStatusValue{
		"approved":    gitlab.LicenseApproved,
		"blacklisted": gitlab.LicenseBlacklisted,
	}

	value, ok := lookup[s]
	if !ok {
		return nil
	}
	return &value
}
