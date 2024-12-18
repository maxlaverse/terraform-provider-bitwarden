package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

type projectOperationFunc func(ctx context.Context, secret models.Project) (*models.Project, error)

func opProjectCreate() secretsManagerOperation {
	return func(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
		return diag.FromErr(projectOperation(ctx, d, bwsClient.CreateProject))
	}
}

func opProjectDelete(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(projectOperation(ctx, d, func(ctx context.Context, project models.Project) (*models.Project, error) {
		return nil, bwsClient.DeleteProject(ctx, project)
	}))
}

func opProjectImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	return []*schema.ResourceData{d}, nil
}

func opProjectRead(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	return diag.FromErr(projectOperation(ctx, d, func(ctx context.Context, projectReq models.Project) (*models.Project, error) {
		project, err := bwsClient.GetProject(ctx, projectReq)
		if project != nil {
			if project.ID != projectReq.ID {
				return nil, errors.New("returned project ID does not match requested project ID")
			}
		}

		return project, err
	}))
}

func opProjectReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	err := projectOperation(ctx, d, func(ctx context.Context, project models.Project) (*models.Project, error) {
		return bwsClient.GetProject(ctx, project)
	})

	if errors.Is(err, models.ErrObjectNotFound) {
		d.SetId("")
		tflog.Warn(ctx, "Project not found, removing from state")
		return diag.Diagnostics{}
	}

	return diag.FromErr(err)
}

func opProjectUpdate(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(projectOperation(ctx, d, bwsClient.EditProject))
}

func projectOperation(ctx context.Context, d *schema.ResourceData, operation projectOperationFunc) error {
	project, err := operation(ctx, projectStructFromData(ctx, d))
	if err != nil {
		return err
	}

	return projectDataFromStruct(ctx, d, project)
}
