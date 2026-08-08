//go:build offline

package schema_definition

import (
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// Schema contract tests lock attribute names and SDKv2 flags (required/optional/
// computed/ForceNew/sensitive and relation fields) so a Framework rewrite can
// be checked against the current public schema surface.

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

func TestSchemaContracts(t *testing.T) {
	cases := []struct {
		name string
		got  map[string]*schema.Schema
		want map[string]attrContract
	}{
		{
			name: "OrgCollection/Resource",
			got:  OrgCollectionSchema(Resource),
			want: map[string]attrContract{
				"id": {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: false, Sensitive: false},
				"member": {Type: schema.TypeSet, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"hide_passwords": {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"id":             {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
					"manage":         {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"read_only":      {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				}},
				"member_group": {Type: schema.TypeSet, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"hide_passwords": {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"id":             {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
					"manage":         {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"read_only":      {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				}},
				"name":            {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				"organization_id": {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "OrgCollection/DataSource",
			got:  OrgCollectionSchema(DataSource),
			want: map[string]attrContract{
				"id": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"member": {Type: schema.TypeSet, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"hide_passwords": {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"id":             {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
					"manage":         {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"read_only":      {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				}},
				"member_group": {Type: schema.TypeSet, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"hide_passwords": {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"id":             {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
					"manage":         {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"read_only":      {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				}},
				"name":            {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"organization_id": {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				"search":          {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, AtLeastOneOf: []string{"search", "id"}},
			},
		},
		{
			name: "AttachmentBase",
			got:  AttachmentSchema(),
			want: map[string]attrContract{
				"file_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"id":        {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"size":      {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"size_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"url":       {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertSchemaContract(t, tc.got, tc.want)
		})
	}
}
