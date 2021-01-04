package gitlab

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabProjectFreezePeriod() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabProjectFreezePeriodCreate,
		Read:   resourceGitlabProjectApprovalRuleRead,
		Update: resourceGitlabProjectApprovalRuleUpdate,
		Delete: resourceGitlabProjectApprovalRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
			},
			"freeze_start": {
				Type:     schema.TypeString,
				Required: true,
			},
			"freeze_end": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cron_timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "UTC",
			},
		},
	}
}

func resourceGitlabProjectFreezePeriodCreate(d *schema.ResourceData, meta interface{}) error {
	project := d.Get("project").(string)

	options := gitlab.CreateFreezePeriodOptions{
		FreezeStart:  gitlab.String(d.Get("freeze_start").(string)),
		FreezeEnd:    gitlab.String(d.Get("freeze_end").(string)),
		CronTimezone: gitlab.String(d.Get("cron_timezone").(string)),
	}

	log.Printf("[DEBUG] Project %s create gitlab project-level freeze period %+v", project, options)

	client := meta.(*gitlab.Client)
	FreezePeriod, _, err := client.FreezePeriods.CreateFreezePeriodOptions(project, &options)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(FreezePeriod.ID))

	return resourceGitlabProjectFreezePeriodRead(d, meta)
}

func resourceGitlabProjectFreezePeriodRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	freezePeriodID, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("%s cannot be converted to int", d.Id())
	}

	log.Printf("[DEBUG] read gitlab FreezePeriod %s/%d", project, freezePeriodID)

	opt := &gitlab.ListFreezePeriodsOptions{
		Page:    1,
		PerPage: 20,
	}

	found := false
	for {
		freezePeriods, resp, err := client.FreezePeriods.ListFreezePeriods(project, opt)
		if err != nil {
			return err
		}
		for _, freezePeriod := range freezePeriods {
			if freezePeriod.ID == freezePeriodID {
				d.Set("id", freezePeriod.ID)
				d.Set("freeze_start", freezePeriod.FreezeStart)
				d.Set("freeze_end", freezePeriod.FreezeEnd)
				d.Set("cron_timezone", freezePeriod.CronTimezone)
				found = true
				break
			}
		}

		if found || resp.CurrentPage >= resp.TotalPages {
			break
		}

		opt.Page = resp.NextPage
	}
	if !found {
		return fmt.Errorf("FreezePeriod %d no longer exists in gitlab", freezePeriodID)
	}

	return nil
}

func resourceGitlabProjectFreezePeriodUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	options := &gitlab.UpdateFreezePeriodOptions{
		FreezeStart:  gitlab.String(d.Get("freeze_start").(string)),
		FreezeEnd:    gitlab.String(d.Get("freeze_end").(string)),
		CronTimezone: gitlab.String(d.Get("cron_timezone").(string)),
	}

	freezePeriodID, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("%s cannot be converted to int", d.Id())
	}

	if d.HasChange("freeze_start") {
		options.FreezeStart = gitlab.String(d.Get("freeze_start").(string))
	}

	if d.HasChange("freeze_end") {
		options.FreezeEnd = gitlab.String(d.Get("freeze_end").(string))
	}

	if d.HasChange("cron_timezone") {
		options.CronTimezone = gitlab.String(d.Get("cron_timezone").(string))
	}

	log.Printf("[DEBUG] update gitlab FreezePeriod %s", d.Id())

	_, _, err = client.FreezePeriods.UpdateFreezePeriodOptions(project, freezePeriodID, options)
	if err != nil {
		return err
	}

	return resourceGitlabProjectFreezePeriodRead(d, meta)
}

func resourceGitlabProjectFreezePeriodDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	log.Printf("[DEBUG] Delete gitlab FreezePeriod %s", d.Id())

	FreezePeriodID, err := strconv.Atoi(d.Id())

	if err != nil {
		return fmt.Errorf("%s cannot be converted to int", d.Id())
	}

	if _, err = client.FreezePeriods.DeleteFreezePeriod(project, FreezePeriodID); err != nil {
		return fmt.Errorf("failed to delete pipeline schedule %q: %w", d.Id(), err)
	}

	return nil
}
