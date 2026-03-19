---
layout: "azuredevops"
page_title: "AzureDevops: azuredevops_pipeline_run"
description: |-
  Manages a pipeline run in Azure DevOps.
---

# azuredevops_pipeline_run

Manages a pipeline run in Azure DevOps. This resource triggers a run on creation.

## Example Usage

```hcl
resource "azuredevops_project" "example" {
  name               = "Example Project"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "azuredevops_git_repository" "example" {
  project_id = azuredevops_project.example.id
  name       = "Example Repo"
  initialization {
    init_type = "Clean"
  }
}

resource "azuredevops_build_definition" "example" {
  project_id = azuredevops_project.example.id
  name       = "Example Pipeline"
  path       = "\\"

  repository {
    repo_type   = "TfsGit"
    repo_id     = azuredevops_git_repository.example.id
    branch_name = azuredevops_git_repository.example.default_branch
    yml_path    = "azure-pipelines.yml"
  }
}

resource "azuredevops_pipeline_run" "example" {
  project_id  = azuredevops_project.example.id
  pipeline_id = azuredevops_build_definition.example.id
  pipeline_parameters = {
    FOO = "BAR"
  }
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) The ID of the project.
* `pipeline_id` - (Required) The ID of the pipeline.
* `pipeline_parameters` - (Optional) The parameters for the pipeline run.
* `pipeline_variables` - (Optional) The variables for the pipeline run.
* `ref_name` - (Optional) The name of the branch or tag for the run. Defaults to `refs/heads/main`.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the pipeline run.
* `run_id` - The ID of the pipeline run.
* `run_number` - The number of the pipeline run.
* `state` - The state of the pipeline run.
* `result` - The result of the pipeline run.
