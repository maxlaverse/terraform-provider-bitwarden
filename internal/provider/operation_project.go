package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/transformation"
)

func opProjectCreate(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwsClient.CreateProject, transformation.ProjectStructFromData, transformation.ProjectDataFromStruct))
}

func opProjectDelete(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, withNilReturn(bwsClient.DeleteProject), transformation.ProjectStructFromData, transformation.ProjectDataFromStruct))
}

func opProjectImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())
	return []*schema.ResourceData{d}, nil
}

func opProjectRead(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	d.SetId(d.Get(schema_definition.AttributeID).(string))
	return diag.FromErr(applyOperation(ctx, d, bwsClient.GetProject, transformation.ProjectStructFromData, transformation.ProjectDataFromStruct))
}

func opProjectReadIgnoreMissing(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return ignoreMissing(ctx, d, applyOperation(ctx, d, bwsClient.GetProject, transformation.ProjectStructFromData, transformation.ProjectDataFromStruct))
}

func opProjectUpdate(ctx context.Context, d *schema.ResourceData, bwsClient bitwarden.SecretsManager) diag.Diagnostics {
	return diag.FromErr(applyOperation(ctx, d, bwsClient.EditProject, transformation.ProjectStructFromData, transformation.ProjectDataFromStruct))
}
