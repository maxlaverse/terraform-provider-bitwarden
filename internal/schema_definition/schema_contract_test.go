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
			name: "ItemBase/Resource",
			got:  ItemBaseSchema(Resource),
			want: map[string]attrContract{
				"collection_ids": {Type: schema.TypeSet, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"creation_date":  {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"deleted_date":   {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"field": {Type: schema.TypeList, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: true, Nested: map[string]attrContract{
					"boolean": {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"hidden":  {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"linked":  {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"name":    {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
					"text":    {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				}},
				"folder_id":       {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"id":              {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: false, Sensitive: false},
				"name":            {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				"notes":           {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: true},
				"organization_id": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: true, Sensitive: false},
				"reprompt":        {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"revision_date":   {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "ItemBase/DataSource",
			got:  ItemBaseSchema(DataSource),
			want: map[string]attrContract{
				"collection_ids": {Type: schema.TypeSet, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"creation_date":  {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"deleted_date":   {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"field": {Type: schema.TypeList, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: true, Nested: map[string]attrContract{
					"boolean": {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"hidden":  {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"linked":  {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"name":    {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
					"text":    {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				}},
				"filter_collection_id":   {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"filter_folder_id":       {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"filter_organization_id": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"folder_id":              {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"id":                     {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"name":                   {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"notes":                  {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: true},
				"organization_id":        {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: true, Sensitive: false},
				"reprompt":               {Type: schema.TypeBool, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"revision_date":          {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"search":                 {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, AtLeastOneOf: []string{"search", "id"}},
			},
		},
		{
			name: "Login/Resource",
			got:  LoginSchema(Resource),
			want: map[string]attrContract{
				"attachments": {Type: schema.TypeList, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"file_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"id":        {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"size":      {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"size_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"url":       {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				}},
				"favorite": {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"password": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: true},
				"totp":     {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: true},
				"uri": {Type: schema.TypeList, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"match": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"value": {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				}},
				"username": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: true},
			},
		},
		{
			name: "Login/DataSource",
			got:  LoginSchema(DataSource),
			want: map[string]attrContract{
				"attachments": {Type: schema.TypeList, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"file_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"id":        {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"size":      {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"size_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"url":       {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				}},
				"favorite":   {Type: schema.TypeBool, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"filter_url": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"password":   {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: true},
				"totp":       {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: true},
				"uri": {Type: schema.TypeList, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"match": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
					"value": {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				}},
				"username": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: true},
			},
		},
		{
			name: "SecureNote/Resource",
			got:  SecureNoteSchema(Resource),
			want: map[string]attrContract{
				"attachments": {Type: schema.TypeList, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"file_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"id":        {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"size":      {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"size_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"url":       {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				}},
				"favorite": {Type: schema.TypeBool, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "SecureNote/DataSource",
			got:  SecureNoteSchema(DataSource),
			want: map[string]attrContract{
				"attachments": {Type: schema.TypeList, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false, Nested: map[string]attrContract{
					"file_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"id":        {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"size":      {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"size_name": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
					"url":       {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				}},
				"favorite": {Type: schema.TypeBool, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "SSHKey/Resource",
			got:  SSHKeySchema(Resource),
			want: map[string]attrContract{
				"key_fingerprint": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: true},
				"private_key":     {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: true},
				"public_key":      {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: true},
			},
		},
		{
			name: "SSHKey/DataSource",
			got:  SSHKeySchema(DataSource),
			want: map[string]attrContract{
				"key_fingerprint": {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: true},
				"private_key":     {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: true},
				"public_key":      {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: true},
			},
		},
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
			name: "OrgGroup/Resource",
			got:  OrgGroupSchema(Resource),
			want: map[string]attrContract{
				"id":              {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: false, Sensitive: false},
				"name":            {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"organization_id": {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "OrgGroup/DataSource",
			got:  OrgGroupSchema(DataSource),
			want: map[string]attrContract{
				"filter_name":     {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, AtLeastOneOf: []string{"filter_name", "id"}},
				"id":              {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"name":            {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"organization_id": {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "OrgMember",
			got:  OrgMemberSchema(),
			want: map[string]attrContract{
				"email":           {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, AtLeastOneOf: []string{"email", "id"}},
				"id":              {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"name":            {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"organization_id": {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "Organization",
			got:  OrganizationSchema(),
			want: map[string]attrContract{
				"id":     {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"name":   {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"search": {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, AtLeastOneOf: []string{"search", "id"}},
			},
		},
		{
			name: "Project/Resource",
			got:  ProjectSchema(Resource),
			want: map[string]attrContract{
				"id":              {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: false, Sensitive: false},
				"name":            {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				"organization_id": {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "Project/DataSource",
			got:  ProjectSchema(DataSource),
			want: map[string]attrContract{
				"id":              {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false},
				"name":            {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"organization_id": {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "Secret/Resource",
			got:  SecretSchema(Resource),
			want: map[string]attrContract{
				"id":              {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: false, Sensitive: false},
				"key":             {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				"note":            {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				"organization_id": {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: false, Sensitive: false},
				"project_id":      {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
				"value":           {Type: schema.TypeString, Required: true, Optional: false, Computed: false, ForceNew: false, Sensitive: false},
			},
		},
		{
			name: "Secret/DataSource",
			got:  SecretSchema(DataSource),
			want: map[string]attrContract{
				"id":              {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, ConflictsWith: []string{"key"}, AtLeastOneOf: []string{"id", "key"}},
				"key":             {Type: schema.TypeString, Required: false, Optional: true, Computed: false, ForceNew: false, Sensitive: false, ConflictsWith: []string{"id"}, AtLeastOneOf: []string{"id", "key"}},
				"note":            {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"organization_id": {Type: schema.TypeString, Required: false, Optional: true, Computed: true, ForceNew: false, Sensitive: false},
				"project_id":      {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
				"value":           {Type: schema.TypeString, Required: false, Optional: false, Computed: true, ForceNew: false, Sensitive: false},
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
