package transformation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func SecretSchemaToObject(_ context.Context, d *schema.ResourceData) models.Secret {
	var secret models.Secret

	secret.ID = d.Id()
	if v, ok := d.Get(schema_definition.AttributeKey).(string); ok {
		secret.Key = v
	}

	if v, ok := d.Get(schema_definition.AttributeValue).(string); ok {
		secret.Value = v
	}

	if v, ok := d.Get(schema_definition.AttributeNote).(string); ok {
		secret.Note = v
	}

	if v, ok := d.Get(schema_definition.AttributeOrganizationID).(string); ok {
		secret.OrganizationID = v
	}

	if v, ok := d.Get(schema_definition.AttributeProjectID).(string); ok {
		secret.ProjectID = v
	}

	return secret
}

func SecretObjectToSchema(_ context.Context, secret *models.Secret, d *schema.ResourceData) error {
	if secret == nil {
		// Secret has been deleted
		return nil
	}

	d.SetId(secret.ID)

	err := d.Set(schema_definition.AttributeKey, secret.Key)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeValue, secret.Value)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeNote, secret.Note)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeOrganizationID, secret.OrganizationID)
	if err != nil {
		return err
	}

	err = d.Set(schema_definition.AttributeProjectID, secret.ProjectID)
	if err != nil {
		return err
	}

	return nil
}
