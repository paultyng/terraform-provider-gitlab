package gitlab

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

func resourceGitlabGroupLabel() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabGroupLabelCreate,
		Read:   resourceGitlabGroupLabelRead,
		Update: resourceGitlabGroupLabelUpdate,
		Delete: resourceGitlabGroupLabelDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGitlabGroupLabelImporter,
		},

		Schema: map[string]*schema.Schema{
			"group": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"color": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceGitlabGroupLabelCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	group := d.Get("group").(string)
	options := &gitlab.CreateGroupLabelOptions{
		Name:  gitlab.String(d.Get("name").(string)),
		Color: gitlab.String(d.Get("color").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		options.Description = gitlab.String(v.(string))
	}

	log.Printf("[DEBUG] create gitlab group label %s", *options.Name)

	label, _, err := client.GroupLabels.CreateGroupLabel(group, options)
	if err != nil {
		return err
	}

	d.SetId(label.Name)

	return resourceGitlabGroupLabelRead(d, meta)
}

func resourceGitlabGroupLabelRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	group := d.Get("group").(string)
	labelName := d.Id()
	log.Printf("[DEBUG] read gitlab group label %s/%s", group, labelName)

	page := 1
	labelsLen := 0
	for page == 1 || labelsLen != 0 {
		labels, _, err := client.GroupLabels.ListGroupLabels(group, &gitlab.ListGroupLabelsOptions{Page: page})
		if err != nil {
			return err
		}
		for _, label := range labels {
			if label.Name == labelName {
				d.Set("description", label.Description)
				d.Set("color", label.Color)
				d.Set("name", label.Name)
				return nil
			}
		}
		labelsLen = len(labels)
		page = page + 1
	}

	log.Printf("[DEBUG] failed to read gitlab label %s/%s", group, labelName)
	d.SetId("")
	return nil
}

func resourceGitlabGroupLabelUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	group := d.Get("group").(string)
	options := &gitlab.UpdateGroupLabelOptions{
		Name:  gitlab.String(d.Get("name").(string)),
		Color: gitlab.String(d.Get("color").(string)),
	}

	if d.HasChange("description") {
		options.Description = gitlab.String(d.Get("description").(string))
	}

	log.Printf("[DEBUG] update gitlab group label %s", d.Id())

	_, _, err := client.GroupLabels.UpdateGroupLabel(group, options)
	if err != nil {
		return err
	}

	return resourceGitlabGroupLabelRead(d, meta)
}

func resourceGitlabGroupLabelDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	group := d.Get("group").(string)
	log.Printf("[DEBUG] Delete gitlab group label %s", d.Id())
	options := &gitlab.DeleteGroupLabelOptions{
		Name: gitlab.String(d.Id()),
	}

	_, err := client.GroupLabels.DeleteGroupLabel(group, options)
	return err
}

func resourceGitlabGroupLabelImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*gitlab.Client)
	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid label id (should be <group ID>.<label name>): %s", d.Id())
	}

	d.SetId(parts[1])
	group, _, err := client.Groups.GetGroup(parts[0], nil)
	if err != nil {
		return nil, err
	}

	if err := d.Set("group", strconv.Itoa(group.ID)); err != nil {
		return nil, err
	}

	err = resourceGitlabGroupLabelRead(d, meta)

	return []*schema.ResourceData{d}, err
}
