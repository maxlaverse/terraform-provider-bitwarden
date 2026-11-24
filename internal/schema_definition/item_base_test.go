//go:build offline

package schema_definition

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensisitiveFieldsAreMarkedAsSensitive(t *testing.T) {
	sensitiveFields := []string{}

	for k, v := range LoginDataSourceSchema().Attributes {
		if v.IsSensitive() {
			sensitiveFields = append(sensitiveFields, k)
		}
	}

	assert.ElementsMatch(t, []string{"notes", "field", "password", "username", "totp"}, sensitiveFields)

	ephemeralSensitive := []string{}
	for k, v := range LoginEphemeralResourceSchema().Attributes {
		if v.IsSensitive() {
			ephemeralSensitive = append(ephemeralSensitive, k)
		}
	}
	assert.ElementsMatch(t, []string{"notes", "field", "password", "username", "totp"}, ephemeralSensitive)

	sshSensitive := []string{}
	for k, v := range SSHKeyEphemeralResourceSchema().Attributes {
		if v.IsSensitive() {
			sshSensitive = append(sshSensitive, k)
		}
	}
	assert.ElementsMatch(t, []string{"notes", "field", "private_key", "public_key", "key_fingerprint"}, sshSensitive)

	noteSensitive := []string{}
	for k, v := range SecureNoteEphemeralResourceSchema().Attributes {
		if v.IsSensitive() {
			noteSensitive = append(noteSensitive, k)
		}
	}
	assert.ElementsMatch(t, []string{"notes", "field"}, noteSensitive)
}

func TestResourceFieldNestedAttributesAreSensitive(t *testing.T) {
	block, ok := LoginResourceSchema().Blocks[AttributeField].(rsschema.ListNestedBlock)
	require.True(t, ok, "field must be a ListNestedBlock")

	for name, attr := range block.NestedObject.Attributes {
		assert.Truef(t, attr.IsSensitive(), "field.%s should be Sensitive (ListNestedBlock cannot mark the list itself)", name)
	}
}

func TestSecretValueAndNoteAreSensitive(t *testing.T) {
	for _, name := range []string{AttributeValue, AttributeNote} {
		attr, ok := SecretResourceSchema().Attributes[name]
		require.Truef(t, ok, "resource %s", name)
		assert.Truef(t, attr.IsSensitive(), "resource %s should be Sensitive", name)

		attr, ok = SecretDataSourceSchema().Attributes[name]
		require.Truef(t, ok, "data source %s", name)
		assert.Truef(t, attr.IsSensitive(), "data source %s should be Sensitive", name)
	}
}

func TestAttachmentContentAndURLAreSensitive(t *testing.T) {
	content, ok := AttachmentResourceSchema().Attributes[AttributeAttachmentContent]
	require.True(t, ok)
	assert.True(t, content.IsSensitive(), "resource content should be Sensitive")

	url, ok := AttachmentResourceSchema().Attributes[AttributeAttachmentURL]
	require.True(t, ok)
	assert.True(t, url.IsSensitive(), "resource url should be Sensitive")

	dsContent, ok := AttachmentDataSourceSchema().Attributes[AttributeAttachmentContent]
	require.True(t, ok)
	assert.True(t, dsContent.IsSensitive(), "data source content should be Sensitive")

	ephemeralContent, ok := AttachmentEphemeralResourceSchema().Attributes[AttributeAttachmentContent]
	require.True(t, ok)
	assert.True(t, ephemeralContent.IsSensitive(), "ephemeral content should be Sensitive")
}

func TestSecretKeyDescriptionIsNotGenericName(t *testing.T) {
	attr, ok := SecretResourceSchema().Attributes[AttributeKey]
	require.True(t, ok)
	assert.Contains(t, attr.GetMarkdownDescription(), "Secrets Manager secret")
	assert.NotEqual(t, "Name.", attr.GetMarkdownDescription())
}

func TestLoginTotpDescriptionIsTheSeed(t *testing.T) {
	attr, ok := LoginResourceSchema().Attributes[AttributeLoginTotp]
	require.True(t, ok)
	assert.Contains(t, attr.GetDescription(), "seed")
	assert.NotEqual(t, "Verification code.", attr.GetDescription())
}

func TestItemResourceOrganizationIDPlanModifiers(t *testing.T) {
	for name, schema := range map[string]rsschema.Schema{
		"login":       LoginResourceSchema(),
		"secure_note": SecureNoteResourceSchema(),
		"ssh_key":     SSHKeyResourceSchema(),
	} {
		attr, ok := schema.Attributes[AttributeOrganizationID].(rsschema.StringAttribute)
		require.Truef(t, ok, "%s organization_id", name)
		require.Lenf(t, attr.PlanModifiers, 2, "%s organization_id", name)
		// RequiresReplace() values are not comparable (function field).
		assert.IsTypef(t, stringplanmodifier.UseStateForUnknown(), attr.PlanModifiers[0], "%s organization_id[0]", name)
		assert.IsTypef(t, stringplanmodifier.RequiresReplace(), attr.PlanModifiers[1], "%s organization_id[1]", name)
	}
}

// Personal items omit organization_id, so Plugin Framework plans it as unknown
// on any other attribute change. RequiresReplace then treats that unknown as a
// possible org move and recreates the item.
func TestOrganizationIDUnknownPlanRequiresReplaceWithoutUseStateForUnknown(t *testing.T) {
	req, resp := unknownOrganizationIDPlan(t)

	stringplanmodifier.RequiresReplace().PlanModifyString(context.Background(), req, resp)

	assert.True(t, resp.RequiresReplace, "unknown organization_id must force replacement without UseStateForUnknown")
}

func TestOrganizationIDUnknownPlanDoesNotReplaceWithItemSchemaModifiers(t *testing.T) {
	attr, ok := LoginResourceSchema().Attributes[AttributeOrganizationID].(rsschema.StringAttribute)
	require.True(t, ok)

	req, resp := unknownStringPlan(t, "")
	for _, m := range attr.PlanModifiers {
		m.PlanModifyString(context.Background(), req, resp)
		req.PlanValue = resp.PlanValue
	}

	assert.False(t, resp.RequiresReplace)
	assert.Equal(t, types.StringValue(""), resp.PlanValue)
}

func TestItemResourceComputedAttributesUseStateForUnknown(t *testing.T) {
	skip := map[string]struct{}{
		AttributeCreationDate: {},
		AttributeDeletedDate:  {},
		AttributeRevisionDate: {},
	}

	for name, schema := range map[string]rsschema.Schema{
		"login":       LoginResourceSchema(),
		"secure_note": SecureNoteResourceSchema(),
		"ssh_key":     SSHKeyResourceSchema(),
	} {
		t.Run(name, func(t *testing.T) {
			for attrName, attr := range schema.Attributes {
				if !attr.IsComputed() {
					continue
				}
				if _, ok := skip[attrName]; ok {
					continue
				}
				assertUseStateForUnknownFirst(t, schema, attrName)
			}
		})
	}
}

func TestOptionalComputedUnknownPlanKeepsState(t *testing.T) {
	login := LoginResourceSchema()

	t.Run("string", func(t *testing.T) {
		attr, ok := login.Attributes[AttributeNotes].(rsschema.StringAttribute)
		require.True(t, ok)

		req, resp := unknownStringPlan(t, "written-in-the-ui")
		for _, m := range attr.PlanModifiers {
			m.PlanModifyString(context.Background(), req, resp)
			req.PlanValue = resp.PlanValue
		}
		assert.Equal(t, types.StringValue("written-in-the-ui"), resp.PlanValue)
	})

	t.Run("bool", func(t *testing.T) {
		attr, ok := login.Attributes[AttributeFavorite].(rsschema.BoolAttribute)
		require.True(t, ok)

		req, resp := unknownBoolPlan(t, true)
		for _, m := range attr.PlanModifiers {
			m.PlanModifyBool(context.Background(), req, resp)
			req.PlanValue = resp.PlanValue
		}
		assert.Equal(t, types.BoolValue(true), resp.PlanValue)
	})

	t.Run("set", func(t *testing.T) {
		collectionIDs, ok := login.Attributes[AttributeCollectionIDs].(rsschema.SetAttribute)
		require.True(t, ok)

		state := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("col-1")})
		req, resp := unknownSetPlan(t, state)
		for _, m := range collectionIDs.PlanModifiers {
			m.PlanModifySet(context.Background(), req, resp)
			req.PlanValue = resp.PlanValue
		}
		assert.Equal(t, state, resp.PlanValue)
	})
}

func assertUseStateForUnknownFirst(t *testing.T, schema rsschema.Schema, name string) {
	t.Helper()

	attr, ok := schema.Attributes[name]
	require.Truef(t, ok, "missing attribute %s", name)

	switch a := attr.(type) {
	case rsschema.StringAttribute:
		require.NotEmptyf(t, a.PlanModifiers, "%s", name)
		assert.IsTypef(t, stringplanmodifier.UseStateForUnknown(), a.PlanModifiers[0], "%s", name)
	case rsschema.BoolAttribute:
		require.NotEmptyf(t, a.PlanModifiers, "%s", name)
		assert.IsTypef(t, boolplanmodifier.UseStateForUnknown(), a.PlanModifiers[0], "%s", name)
	case rsschema.SetAttribute:
		require.NotEmptyf(t, a.PlanModifiers, "%s", name)
		assert.IsTypef(t, setplanmodifier.UseStateForUnknown(), a.PlanModifiers[0], "%s", name)
	case rsschema.ListNestedAttribute:
		require.NotEmptyf(t, a.PlanModifiers, "%s", name)
		assert.IsTypef(t, listplanmodifier.UseStateForUnknown(), a.PlanModifiers[0], "%s", name)
	default:
		t.Fatalf("%s: unexpected attribute type %T", name, attr)
	}
}

func unknownOrganizationIDPlan(t *testing.T) (planmodifier.StringRequest, *planmodifier.StringResponse) {
	return unknownStringPlan(t, "")
}

func unknownStringPlan(t *testing.T, state string) (planmodifier.StringRequest, *planmodifier.StringResponse) {
	t.Helper()

	raw := existingItemRaw()
	req := planmodifier.StringRequest{
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringUnknown(),
		StateValue:  types.StringValue(state),
		Plan:        tfsdk.Plan{Raw: raw},
		State:       tfsdk.State{Raw: raw},
	}
	return req, &planmodifier.StringResponse{PlanValue: req.PlanValue}
}

func unknownBoolPlan(t *testing.T, state bool) (planmodifier.BoolRequest, *planmodifier.BoolResponse) {
	t.Helper()

	raw := existingItemRaw()
	req := planmodifier.BoolRequest{
		ConfigValue: types.BoolNull(),
		PlanValue:   types.BoolUnknown(),
		StateValue:  types.BoolValue(state),
		Plan:        tfsdk.Plan{Raw: raw},
		State:       tfsdk.State{Raw: raw},
	}
	return req, &planmodifier.BoolResponse{PlanValue: req.PlanValue}
}

func unknownSetPlan(t *testing.T, state types.Set) (planmodifier.SetRequest, *planmodifier.SetResponse) {
	t.Helper()

	raw := existingItemRaw()
	req := planmodifier.SetRequest{
		ConfigValue: types.SetNull(types.StringType),
		PlanValue:   types.SetUnknown(types.StringType),
		StateValue:  state,
		Plan:        tfsdk.Plan{Raw: raw},
		State:       tfsdk.State{Raw: raw},
	}
	return req, &planmodifier.SetResponse{PlanValue: req.PlanValue}
}

func existingItemRaw() tftypes.Value {
	// Non-null resource plan/state so RequiresReplace does not treat this as create/destroy.
	return tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{"id": tftypes.String},
	}, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "item-1"),
	})
}
