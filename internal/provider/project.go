package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

type projectOperationFunc func(ctx context.Context, secret models.Project) (*models.Project, error)

func projectRead(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
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

func projectOperation(ctx context.Context, d *schema.ResourceData, operation projectOperationFunc) error {
	project, err := operation(ctx, projectStructFromData(ctx, d))
	if err != nil {
		return err
	}

	return projectDataFromStruct(ctx, d, project)
}

func projectStructFromData(_ context.Context, d *schema.ResourceData) models.Project {
	var project models.Project

	project.ID = d.Id()
	if v, ok := d.Get(attributeName).(string); ok {
		project.Name = v
	}

	if v, ok := d.Get(attributeOrganizationID).(string); ok {
		project.OrganizationID = v
	}

	return project
}

func projectDataFromStruct(_ context.Context, d *schema.ResourceData, project *models.Project) error {
	if project == nil {
		// Project has been deleted
		return nil
	}

	d.SetId(project.ID)

	err := d.Set(attributeName, project.Name)
	if err != nil {
		return err
	}

	err = d.Set(attributeOrganizationID, project.OrganizationID)
	if err != nil {
		return err
	}

	return nil
}
