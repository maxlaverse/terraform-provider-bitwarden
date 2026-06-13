package transformation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

func UserObjectToSchema(ctx context.Context, obj *models.User, d *schema.ResourceData) error {
	if obj == nil {
		// Object has been deleted
		return nil
	}

	d.SetId(obj.ID)

	return d.Set(schema_definition.AttributeEmail, obj.Email)
}

func UserToObject(ctx context.Context, d *schema.ResourceData) models.User {
	var obj models.User

	obj.ID = d.Id()
	if v, ok := d.Get(schema_definition.AttributeEmail).(string); ok {
		obj.Email = v
	}

	return obj
}
