//go:build offline

package provider

import (
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

type attrContract struct {
	Type          schema.ValueType
	Required      bool
	Optional      bool
	Computed      bool
	ForceNew      bool
	Sensitive     bool
	ConflictsWith []string
	AtLeastOneOf  []string
	RequiredWith  []string
	ComputedWhen  []string
	Nested        map[string]attrContract
}

func assertSchemaContract(t *testing.T, got map[string]*schema.Schema, want map[string]attrContract) {
	t.Helper()

	gotKeys := make([]string, 0, len(got))
	for k := range got {
		gotKeys = append(gotKeys, k)
	}
	wantKeys := make([]string, 0, len(want))
	for k := range want {
		wantKeys = append(wantKeys, k)
	}
	sort.Strings(gotKeys)
	sort.Strings(wantKeys)
	assert.Equal(t, wantKeys, gotKeys, "attribute names")

	for name, expected := range want {
		actual := got[name]
		if !assert.NotNil(t, actual, "missing attribute %q", name) {
			continue
		}
		assert.Equal(t, expected.Type, actual.Type, "%s.Type", name)
		assert.Equal(t, expected.Required, actual.Required, "%s.Required", name)
		assert.Equal(t, expected.Optional, actual.Optional, "%s.Optional", name)
		assert.Equal(t, expected.Computed, actual.Computed, "%s.Computed", name)
		assert.Equal(t, expected.ForceNew, actual.ForceNew, "%s.ForceNew", name)
		assert.Equal(t, expected.Sensitive, actual.Sensitive, "%s.Sensitive", name)
		assert.Equal(t, expected.ConflictsWith, actual.ConflictsWith, "%s.ConflictsWith", name)
		assert.Equal(t, expected.AtLeastOneOf, actual.AtLeastOneOf, "%s.AtLeastOneOf", name)
		assert.Equal(t, expected.RequiredWith, actual.RequiredWith, "%s.RequiredWith", name)
		assert.Equal(t, expected.ComputedWhen, actual.ComputedWhen, "%s.ComputedWhen", name)

		if expected.Nested == nil {
			continue
		}
		res, ok := actual.Elem.(*schema.Resource)
		if !assert.True(t, ok && res != nil, "%s.Elem should be *schema.Resource", name) {
			continue
		}
		assertSchemaContract(t, res.Schema, expected.Nested)
	}
}

func TestProviderAndAttachmentSchemaContracts(t *testing.T) {
	// Attachment resource/data source schemas stay on SDKv2 during the mux
	// migration and must remain backward compatible (hashes, ForceNew, conflicts).
	// Provider schema mux alignment is covered by TestProviderSchemaValidity.
	cases := []struct {
		name string
		got  map[string]*schema.Schema
		want map[string]attrContract
	}{
		{
			name: "ResourceAttachment",
			got:  resourceAttachment().Schema,
			want: map[string]attrContract{
				"content":   {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: true, Sensitive: false, ConflictsWith: []string{"file"}, AtLeastOneOf: []string{"file"}, RequiredWith: []string{"content"}},
				"file":      {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: true, Sensitive: false, ConflictsWith: []string{"content"}, AtLeastOneOf: []string{"content"}},
				"file_name": {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: true, Sensitive: false, ConflictsWith: []string{"file"}, AtLeastOneOf: []string{"file"}, RequiredWith: []string{"content"}, ComputedWhen: []string{"file"}},
				"id":        {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"item_id":   {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: true, Sensitive: false},
				"size":      {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"size_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"url":       {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "DataSourceAttachment",
			got:  dataSourceAttachment().Schema,
			want: map[string]attrContract{
				"content": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"id":      {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				"item_id": {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertSchemaContract(t, tc.got, tc.want)
		})
	}
}
