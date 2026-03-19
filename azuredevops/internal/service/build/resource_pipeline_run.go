package build

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/client"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils/converter"
)

// ResourcePipelineRun schema and implementation for pipeline run resource
func ResourcePipelineRun() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePipelineRunCreate,
		ReadContext:   resourcePipelineRunRead,
		DeleteContext: resourcePipelineRunDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pipeline_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"pipeline_parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pipeline_variables": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ref_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "refs/heads/main",
			},
			"run_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"run_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"result": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePipelineRunCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	pipelineID := d.Get("pipeline_id").(int)

	runParams := &pipelines.RunPipelineParameters{}

	if v, ok := d.GetOk("pipeline_parameters"); ok {
		params := make(map[string]string)
		for k, val := range v.(map[string]interface{}) {
			params[k] = val.(string)
		}
		runParams.TemplateParameters = &params
	}

	if v, ok := d.GetOk("pipeline_variables"); ok {
		vars := make(map[string]pipelines.Variable)
		for k, val := range v.(map[string]interface{}) {
			vars[k] = pipelines.Variable{
				Value: converter.String(val.(string)),
			}
		}
		runParams.Variables = &vars
	}

	if v, ok := d.GetOk("ref_name"); ok {
		runParams.Resources = &pipelines.RunResourcesParameters{
			Repositories: &map[string]pipelines.RepositoryResourceParameters{
				"self": {
					RefName: converter.String(v.(string)),
				},
			},
		}
	}

	run, err := clients.PipelinesClient.RunPipeline(ctx, pipelines.RunPipelineArgs{
		Project:       &projectID,
		PipelineId:    &pipelineID,
		RunParameters: runParams,
	})
	if err != nil {
		return diag.Errorf("Triggering pipeline run: %+v", err)
	}

	d.SetId(strconv.Itoa(*run.Id))
	return resourcePipelineRunRead(ctx, d, m)
}

func resourcePipelineRunRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clients := m.(*client.AggregatedClient)
	projectID := d.Get("project_id").(string)
	pipelineID := d.Get("pipeline_id").(int)
	runID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Parsing run ID: %+v", err)
	}

	run, err := clients.PipelinesClient.GetRun(ctx, pipelines.GetRunArgs{
		Project:    &projectID,
		PipelineId: &pipelineID,
		RunId:      &runID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Reading pipeline run: %+v", err)
	}

	d.Set("run_id", *run.Id)
	if run.Name != nil {
		d.Set("run_number", *run.Name)
	}
	if run.State != nil {
		d.Set("state", string(*run.State))
	}
	if run.Result != nil {
		d.Set("result", string(*run.Result))
	}

	return nil
}

func resourcePipelineRunDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Runs are one-off entities that already exist in history.
	// We don't need to do anything to "delete" the run record from Azure DevOps.
	d.SetId("")
	return nil
}
