package bwcli

import "github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"

func FilterObjectsByType(objs []models.Object, itemType models.ItemType) []models.Object {
	if itemType == 0 {
		return objs
	}

	filtered := make([]models.Object, 0, len(objs))
	for _, obj := range objs {
		if obj.Type == itemType {
			filtered = append(filtered, obj)
		}
	}
	return filtered
}
