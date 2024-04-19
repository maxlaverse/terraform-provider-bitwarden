package bw

func FilterObjectsByType(objs []Object, itemType ItemType) []Object {
	if itemType == 0 {
		return objs
	}

	filtered := make([]Object, 0, len(objs))
	for _, obj := range objs {
		if obj.Type == itemType {
			filtered = append(filtered, obj)
		}
	}
	return filtered
}
