package transformation

import (
	"context"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func ProjectSchemaToObject(_ context.Context, d AttrData) models.Project {
	var project models.Project

	project.ID = d.Id()
	if v, ok := d.Get(schema_definition.AttributeName).(string); ok {
		project.Name = v
	}

	if v, ok := d.Get(schema_definition.AttributeOrganizationID).(string); ok {
		project.OrganizationID = v
	}

	return project
}

func ProjectObjectToSchema(_ context.Context, project *models.Project, d AttrData) error {
	if project == nil {
		// Project has been deleted
		return nil
	}

	d.SetId(project.ID)

	err := d.Set(schema_definition.AttributeName, project.Name)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeOrganizationID, project.OrganizationID)
	if err != nil {
		return err
	}

	return nil
}
