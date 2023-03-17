package waypoint

import (
	"context"
	"fmt"
	"github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	gen "github.com/hashicorp-dev-advocates/waypoint-client/pkg/waypoint"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRunnerProfile() *schema.Resource {
	return &schema.Resource{
		Description: "Runner profile resource to configure Waypoint runners.",

		CreateContext: resourceRunnerProfileCreate,
		ReadContext:   resourceRunnerProfileRead,
		DeleteContext: resourceRunnerProfileDelete,
		UpdateContext: resourceRunnerProfileUpdate,

		Schema: map[string]*schema.Schema{
			"profile_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the runner profile",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Computed ID of runner profile.",
			},
			"oci_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "oci_url is the OCI image that will be used to boot the on demand runner.",
			},
			"plugin_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Plugin type for runner i.e docker / kubernetes / aws-ecs.",
			},
			"plugin_config": {
				Type:        schema.TypeString, // Under the hood the type is []byte
				Optional:    true,
				Description: "plugin config is the configuration for the plugin that is created. It is usually HCL and is decoded like the other plugins, and is plugin specific.",
			},
			"plugin_config_format": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "config format specifies the format of plugin_config.",
			},
			"default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if this runner profile is the default for any new projects",
			},
			"target_runner_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The ID of the target runner for this profile.",
				ConflictsWith: []string{"target_runner_labels"},
			},
			"target_runner_labels": {
				Type:          schema.TypeMap,
				Optional:      true,
				Description:   "A map of labels on target runners",
				ConflictsWith: []string{"target_runner_id"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Any env vars that should be exposed to the on demand runner.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRunnerProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	wp := m.(*WaypointClient).conn

	runnerConfig := client.DefaultRunnerConfig()
	runnerConfig.Name = d.Get("profile_name").(string)

	if ociUrl, ok := d.Get("oci_url").(string); ok {
		runnerConfig.OciUrl = ociUrl
	}

	if pluginType, ok := d.Get("plugin_type").(string); ok {
		runnerConfig.PluginType = pluginType
	}

	if pluginConfig, ok := d.Get("plugin_config").(string); ok {
		runnerConfig.PluginConfig = []byte(pluginConfig)
	}

	if pluginConfigFormat, ok := d.Get("plugin_config_format").(int); ok {
		runnerConfig.ConfigFormat = pluginConfigFormat
	}

	if defaultProfile, ok := d.Get("default").(bool); ok {
		runnerConfig.Default = defaultProfile
	}

	tRId := d.Get("target_runner_id").(string)
	if len(tRId) > 0 {

		if targetRunnerId, ok := d.Get("target_runner_id").(string); ok {
			runnerConfig.TargetRunner = &gen.Ref_Runner{Target: &gen.Ref_Runner_Id{Id: &gen.Ref_RunnerId{Id: targetRunnerId}}}
		}
	}

	tRL := d.Get("target_runner_labels").(map[string]interface{})

	if len(tRL) > 0 {

		if targetRunnerLabels, ok := d.Get("target_runner_labels").(map[string]interface{}); ok {
			labels := make(map[string]string)

			for k, v := range targetRunnerLabels {
				strKey := fmt.Sprintf("%v", k)
				strValue := fmt.Sprintf("%v", v)
				labels[strKey] = strValue
			}

			runnerConfig.TargetRunner.Target = &gen.Ref_Runner_Labels{
				Labels: &gen.Ref_RunnerLabels{
					Labels: labels,
				}}
		}
	}

	runnerVariables := make(map[string]string)
	if environmentVariables, ok := d.Get("environment_variables").(map[string]interface{}); ok {

		for k, v := range environmentVariables {
			strKey := fmt.Sprintf("%v", k)
			strValue := fmt.Sprintf("%v", v)
			runnerVariables[strKey] = strValue
		}

		runnerConfig.EnvironmentVariables = runnerVariables

	}

	runnerProfile, err := wp.CreateRunnerProfile(context.TODO(), runnerConfig)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(runnerProfile.Config.Id)

	tflog.Trace(ctx, "created a resource")

	return resourceRunnerProfileRead(ctx, d, m)
}

func resourceRunnerProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	wp := m.(*WaypointClient).conn

	profileId := d.Get("id").(string)

	getRunnerProfile, err := wp.GetRunnerProfile(context.TODO(), profileId)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(profileId)
	d.Set("profile_name", getRunnerProfile.Config.Name)
	d.Set("oci_url", getRunnerProfile.Config.OciUrl)
	d.Set("plugin_type", getRunnerProfile.Config.PluginType)
	d.Set("plugin_config", getRunnerProfile.Config.PluginConfig)
	d.Set("plugin_config_format", getRunnerProfile.Config.ConfigFormat)
	d.Set("default", getRunnerProfile.Config.Default)

	switch getRunnerProfile.Config.TargetRunner.Target.(type) {
	case *gen.Ref_Runner_Labels:
		d.Set("target_runner_labels", getRunnerProfile.Config.TargetRunner.Target.(*gen.Ref_Runner_Labels).Labels.Labels)
	case *gen.Ref_Runner_Id:
		d.Set("target_runner_id", getRunnerProfile.Config.TargetRunner.Target.(*gen.Ref_Runner_Id).Id.Id)

	}

	d.Set("environment_variables", getRunnerProfile.Config.EnvironmentVariables)

	return nil
}

func resourceRunnerProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	wp := m.(*WaypointClient).conn

	runnerConfig := client.DefaultRunnerConfig()
	runnerConfig.Name = d.Get("profile_name").(string)
	runnerConfig.Id = d.Get("id").(string)

	if ociUrl, ok := d.Get("oci_url").(string); ok {
		runnerConfig.OciUrl = ociUrl
	}

	if pluginType, ok := d.Get("plugin_type").(string); ok {
		runnerConfig.PluginType = pluginType
	}

	if pluginConfig, ok := d.Get("plugin_config").(string); ok {
		runnerConfig.PluginConfig = []byte(pluginConfig)
	}

	if pluginConfigFormat, ok := d.Get("plugin_config_format").(int); ok {
		runnerConfig.ConfigFormat = pluginConfigFormat
	}

	if defaultProfile, ok := d.Get("default").(bool); ok {
		runnerConfig.Default = defaultProfile
	}

	tRId := d.Get("target_runner_id").(string)
	if len(tRId) > 0 {

		if targetRunnerId, ok := d.Get("target_runner_id").(string); ok {
			runnerConfig.TargetRunner = &gen.Ref_Runner{Target: &gen.Ref_Runner_Id{Id: &gen.Ref_RunnerId{Id: targetRunnerId}}}
		}
	}

	tRL := d.Get("target_runner_labels").(map[string]interface{})

	if len(tRL) > 0 {

		if targetRunnerLabels, ok := d.Get("target_runner_labels").(map[string]interface{}); ok {
			labels := make(map[string]string)

			for k, v := range targetRunnerLabels {
				strKey := fmt.Sprintf("%v", k)
				strValue := fmt.Sprintf("%v", v)
				labels[strKey] = strValue
			}

			runnerConfig.TargetRunner.Target = &gen.Ref_Runner_Labels{
				Labels: &gen.Ref_RunnerLabels{
					Labels: labels,
				}}
		}
	}

	if environmentVariables, ok := d.Get("environment_variables").(map[string]interface{}); ok {
		runnerVariables := make(map[string]string)

		for k, v := range environmentVariables {
			strKey := fmt.Sprintf("%v", k)
			strValue := fmt.Sprintf("%v", v)
			runnerVariables[strKey] = strValue
		}

		runnerConfig.EnvironmentVariables = runnerVariables

	}

	runnerProfile, err := wp.CreateRunnerProfile(context.TODO(), runnerConfig)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(runnerProfile.Config.Id)

	tflog.Trace(ctx, "created a resource")

	return resourceRunnerProfileRead(ctx, d, m)
}

func resourceRunnerProfileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	wp := m.(*WaypointClient).conn

	runnerConfig := client.DefaultRunnerConfig()
	runnerConfig.Name = "RESOURCE DELETED"
	runnerConfig.Id = d.Get("id").(string)
	runnerConfig.OciUrl = "RESOURCE DELETED"
	runnerConfig.PluginType = "RESOURCE DELETED"

	runnerProfile, err := wp.CreateRunnerProfile(context.TODO(), runnerConfig)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(runnerProfile.Config.Id)

	tflog.Trace(ctx, "updated a resource")

	return resourceRunnerProfileRead(ctx, d, m)
}
