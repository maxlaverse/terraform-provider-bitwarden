package transformation

import (
	"context"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMapDataNil(t *testing.T) {
	t.Parallel()

	m := NewMapData(nil)
	require.NotNil(t, m.Values())
	assert.Empty(t, m.Values())
	assert.Empty(t, m.Id())
}

func TestMapDataGetSetId(t *testing.T) {
	t.Parallel()

	m := NewMapData(map[string]interface{}{
		schema_definition.AttributeName: "inbox",
	})
	assert.Equal(t, "inbox", m.Get(schema_definition.AttributeName))
	assert.Nil(t, m.Get("missing"))

	m.SetId("folder-1")
	assert.Equal(t, "folder-1", m.Id())

	require.NoError(t, m.Set(schema_definition.AttributeName, "archive"))
	assert.Equal(t, "archive", m.Values()[schema_definition.AttributeName])
}

func TestMapDataGetOk(t *testing.T) {
	t.Parallel()

	m := NewMapData(map[string]interface{}{
		"present":      "value",
		"emptyString":  "",
		"trueBool":     true,
		"falseBool":    false,
		"zeroInt":      0,
		"nonZeroInt":   3,
		"emptySlice":   []interface{}{},
		"slice":        []interface{}{"a"},
		"emptyStrings": []string{},
		"strings":      []string{"b"},
		"nilValue":     nil,
	})

	v, ok := m.GetOk("present")
	assert.True(t, ok)
	assert.Equal(t, "value", v)

	_, ok = m.GetOk("missing")
	assert.False(t, ok)

	_, ok = m.GetOk("emptyString")
	assert.False(t, ok)

	_, ok = m.GetOk("falseBool")
	assert.False(t, ok)

	v, ok = m.GetOk("trueBool")
	assert.True(t, ok)
	assert.Equal(t, true, v)

	_, ok = m.GetOk("zeroInt")
	assert.False(t, ok)

	v, ok = m.GetOk("nonZeroInt")
	assert.True(t, ok)
	assert.Equal(t, 3, v)

	_, ok = m.GetOk("emptySlice")
	assert.False(t, ok)

	v, ok = m.GetOk("slice")
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"a"}, v)

	_, ok = m.GetOk("emptyStrings")
	assert.False(t, ok)

	v, ok = m.GetOk("strings")
	assert.True(t, ok)
	assert.Equal(t, []string{"b"}, v)

	_, ok = m.GetOk("nilValue")
	assert.False(t, ok)
}

func TestMapDataFolderRoundTrip(t *testing.T) {
	t.Parallel()

	attr := NewMapData(map[string]interface{}{
		schema_definition.AttributeName: "Personal",
	})
	attr.SetId("")

	obj := SchemaToFolderObject(context.Background(), attr)
	assert.Equal(t, "Personal", obj.Name)
	assert.Equal(t, models.ObjectTypeFolder, obj.Object)

	created := &models.Folder{
		ID:     "folder-abc",
		Name:   "Personal",
		Object: models.ObjectTypeFolder,
	}
	require.NoError(t, FolderObjectToSchema(context.Background(), created, attr))
	assert.Equal(t, "folder-abc", attr.Id())
	assert.Equal(t, "Personal", attr.Values()[schema_definition.AttributeName])
}

func TestMapDataListOptionsFromData(t *testing.T) {
	t.Parallel()

	attr := NewMapData(map[string]interface{}{
		schema_definition.AttributeFilterSearch: "vault",
		schema_definition.AttributeFilterURL:    "",
	})
	opts := ListOptionsFromData(attr)
	assert.Len(t, opts, 1)
}

func TestMapDataItemCollectionIDsRoundTrip(t *testing.T) {
	t.Parallel()

	attr := NewMapData(map[string]interface{}{
		schema_definition.AttributeName:          "login",
		schema_definition.AttributeCollectionIDs: []string{"col-1", "col-2"},
		schema_definition.AttributeFavorite:      true,
	})
	attr.SetId("item-1")

	obj := ItemSchemaToObject(models.ItemTypeLogin)(context.Background(), attr)
	assert.Equal(t, "item-1", obj.ID)
	assert.Equal(t, []string{"col-1", "col-2"}, obj.CollectionIds)
	assert.True(t, obj.Favorite)

	require.NoError(t, ItemObjectToSchema(context.Background(), &obj, attr))
	assert.Equal(t, "item-1", attr.Id())
	assert.Equal(t, []string{"col-1", "col-2"}, attr.Values()[schema_definition.AttributeCollectionIDs])
}
