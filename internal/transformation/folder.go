package transformation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func FolderObjectToSchema(ctx context.Context, obj *models.Folder, d *schema.ResourceData) error {
	if obj == nil {
		// Object has been deleted
		return nil
	}

	d.SetId(obj.ID)

	err := d.Set(schema_definition.AttributeName, obj.Name)
	if err != nil {
		return err
	}

	return nil
}

func SchemaToFolderObject(ctx context.Context, d *schema.ResourceData) models.Folder {
	var obj models.Folder

	obj.ID = d.Id()
	if v, ok := d.Get(schema_definition.AttributeName).(string); ok {
		obj.Name = v
	}

	obj.Object = models.ObjectTypeFolder
	return obj
}
