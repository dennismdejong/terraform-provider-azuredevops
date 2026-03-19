package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/acceptancetests/testutils"
)

func TestAccPipelineRun_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	pipelineName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclPipelineRunBasic(projectName, pipelineName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("azuredevops_pipeline_run.run", "project_id"),
					resource.TestCheckResourceAttrSet("azuredevops_pipeline_run.run", "pipeline_id"),
					resource.TestCheckResourceAttrSet("azuredevops_pipeline_run.run", "run_id"),
					resource.TestCheckResourceAttr("azuredevops_pipeline_run.run", "pipeline_parameters.FOO", "BAR"),
				),
			},
		},
	})
}

func hclPipelineRunBasic(projectName, pipelineName string) string {
	return fmt.Sprintf(`
resource "azuredevops_project" "project" {
  name               = "%s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "azuredevops_git_repository" "repo" {
  project_id = azuredevops_project.project.id
  name       = "Sample Repo"
  initialization {
    init_type = "Clean"
  }
}

resource "azuredevops_build_definition" "pipeline" {
  project_id = azuredevops_project.project.id
  name       = "%s"
  path       = "\\"

  repository {
    repo_type   = "TfsGit"
    repo_id     = azuredevops_git_repository.repo.id
    branch_name = azuredevops_git_repository.repo.default_branch
    yml_path    = "azure-pipelines.yml"
  }
}

resource "azuredevops_git_repository_file" "file" {
  repository_id = azuredevops_git_repository.repo.id
  file          = "azure-pipelines.yml"
  content       = "parameters:\n- name: FOO\n  type: string\n  default: 'BAZ'\nsteps:\n- script: echo $(FOO)"
  branch        = azuredevops_git_repository.repo.default_branch
  commit_message = "Add pipeline"
  overwrite_on_create = true
}

resource "azuredevops_pipeline_run" "run" {
  project_id  = azuredevops_project.project.id
  pipeline_id = azuredevops_build_definition.pipeline.id
  pipeline_parameters = {
    FOO = "BAR"
  }
  depends_on = [azuredevops_git_repository_file.file]
}
`, projectName, pipelineName)
}
