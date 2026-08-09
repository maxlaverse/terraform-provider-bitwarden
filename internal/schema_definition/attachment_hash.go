package schema_definition

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Attachment file/content under Protocol 6.
//
// State stores the configured path (`file`) or raw body (`content`). SDKv2
// stored a SHA1 via StateFunc. Hashed* types compare by content digest:
//
//   - Keep: two paths with identical bytes do not force replace (ForceNew-by-content).
//   - Drop later: leftover SDKv2 SHA1 values in state compare equal to the
//     matching path/body. That is only isSHA1Hex plus the SHA1 branch in
//     digestOf*Side.
//
// Plan modifiers set RequiresReplace when digests differ.

// FileSha1Sum returns the SHA1 hex digest of the contents of the file at
// filepath.
func FileSha1Sum(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err = io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// ContentSha1Sum returns the SHA1 hex digest of content.
func ContentSha1Sum(content string) (string, error) {
	hash := sha1.New()
	if _, err := hash.Write([]byte(content)); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// isSHA1Hex reports whether s looks like a 40-char hex SHA1, as SDKv2 StateFunc
// wrote into state. Drop with the SHA1 branches in digestOf*Side once those
// leftover states have been rewritten to a path/body.
//
// Ambiguous: a configured `content` or `file` that is itself 40 hex chars is
// treated as a digest, not hashed, and can force an extra replace on upgrade.
func isSHA1Hex(s string) bool {
	if len(s) != sha1.Size*2 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// digestOfFileSide hashes a filesystem path, or returns raw as-is when it is a
// leftover SDKv2 SHA1 (the droppable branch).
func digestOfFileSide(raw string) (string, error) {
	if isSHA1Hex(raw) {
		return raw, nil
	}
	return FileSha1Sum(raw)
}

// digestOfContentSide hashes raw content, or returns raw as-is when it is a
// leftover SDKv2 SHA1 (the droppable branch).
func digestOfContentSide(raw string) (string, error) {
	if isSHA1Hex(raw) {
		return raw, nil
	}
	return ContentSha1Sum(raw)
}

// --- "file" attribute ---

var (
	_ basetypes.StringTypable                    = HashedFileType{}
	_ basetypes.StringValuableWithSemanticEquals = HashedFileValue{}
)

// HashedFileType is the CustomType of the attachment "file" attribute.
type HashedFileType struct {
	basetypes.StringType
}

func (t HashedFileType) Equal(o attr.Type) bool {
	other, ok := o.(HashedFileType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t HashedFileType) String() string {
	return "HashedFileType"
}

func (t HashedFileType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return HashedFileValue{StringValue: in}, nil
}

func (t HashedFileType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t HashedFileType) ValueType(_ context.Context) attr.Value {
	return HashedFileValue{}
}

// HashedFileValue holds a configured file path (or a legacy SHA1 digest from
// SDKv2 state).
type HashedFileValue struct {
	basetypes.StringValue
}

// NewHashedFileValue builds a HashedFileValue from a raw string.
func NewHashedFileValue(raw string) HashedFileValue {
	return HashedFileValue{StringValue: types.StringValue(raw)}
}

func (v HashedFileValue) Equal(o attr.Value) bool {
	other, ok := o.(HashedFileValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v HashedFileValue) Type(_ context.Context) attr.Type {
	return HashedFileType{}
}

// StringSemanticEquals compares two file sides by content digest (path or
// legacy hash).
func (v HashedFileValue) StringSemanticEquals(_ context.Context, other basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	otherValue, ok := other.(HashedFileValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks on the attachment \"file\" attribute. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Expected Value Type: %T\nGot Value Type: %T", v, other),
		)
		return false, diags
	}

	a, b := v.ValueString(), otherValue.ValueString()
	if a == b {
		return true, diags
	}

	digestA, err := digestOfFileSide(a)
	if err != nil {
		diags.AddError("Unable to Compute Attachment File Hash", fmt.Sprintf("unable to read file %q to compute its hash: %s", a, err))
		return false, diags
	}
	digestB, err := digestOfFileSide(b)
	if err != nil {
		diags.AddError("Unable to Compute Attachment File Hash", fmt.Sprintf("unable to read file %q to compute its hash: %s", b, err))
		return false, diags
	}
	return digestA == digestB, diags
}

// --- "content" attribute ---

var (
	_ basetypes.StringTypable                    = HashedContentType{}
	_ basetypes.StringValuableWithSemanticEquals = HashedContentValue{}
)

// HashedContentType is the CustomType of the attachment "content" attribute.
type HashedContentType struct {
	basetypes.StringType
}

func (t HashedContentType) Equal(o attr.Type) bool {
	other, ok := o.(HashedContentType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t HashedContentType) String() string {
	return "HashedContentType"
}

func (t HashedContentType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return HashedContentValue{StringValue: in}, nil
}

func (t HashedContentType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t HashedContentType) ValueType(_ context.Context) attr.Value {
	return HashedContentValue{}
}

// HashedContentValue holds configured raw content (or a legacy SHA1 digest).
type HashedContentValue struct {
	basetypes.StringValue
}

// NewHashedContentValue builds a HashedContentValue from a raw string.
func NewHashedContentValue(raw string) HashedContentValue {
	return HashedContentValue{StringValue: types.StringValue(raw)}
}

func (v HashedContentValue) Equal(o attr.Value) bool {
	other, ok := o.(HashedContentValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v HashedContentValue) Type(_ context.Context) attr.Type {
	return HashedContentType{}
}

// StringSemanticEquals compares two content sides by digest (raw content or
// legacy hash).
func (v HashedContentValue) StringSemanticEquals(_ context.Context, other basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	otherValue, ok := other.(HashedContentValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks on the attachment \"content\" attribute. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Expected Value Type: %T\nGot Value Type: %T", v, other),
		)
		return false, diags
	}

	a, b := v.ValueString(), otherValue.ValueString()
	if a == b {
		return true, diags
	}

	digestA, err := digestOfContentSide(a)
	if err != nil {
		diags.AddError("Unable to Compute Attachment Content Hash", err.Error())
		return false, diags
	}
	digestB, err := digestOfContentSide(b)
	if err != nil {
		diags.AddError("Unable to Compute Attachment Content Hash", err.Error())
		return false, diags
	}
	return digestA == digestB, diags
}

// --- Plan modifiers ---

// AttachmentFilePlanModifiers mirrors SDKv2 ForceNew based on file contents.
func AttachmentFilePlanModifiers() []planmodifier.String {
	return []planmodifier.String{hashedFilePlanModifier{}}
}

// AttachmentContentPlanModifiers mirrors SDKv2 ForceNew based on content.
func AttachmentContentPlanModifiers() []planmodifier.String {
	return []planmodifier.String{hashedContentPlanModifier{}}
}

func planModifyHashedAttribute(req planmodifier.StringRequest, resp *planmodifier.StringResponse, semanticEquals func(configRaw, stateRaw string) (bool, diag.Diagnostics)) {
	if req.ConfigValue.IsUnknown() {
		resp.PlanValue = types.StringUnknown()
		return
	}

	if req.ConfigValue.IsNull() {
		resp.PlanValue = types.StringNull()
		if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
			resp.RequiresReplace = true
		}
		return
	}

	// Protocol 6: planned value must match config.
	resp.PlanValue = req.ConfigValue

	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	equal, diags := semanticEquals(req.ConfigValue.ValueString(), req.StateValue.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !equal {
		resp.RequiresReplace = true
	}
}

type hashedFilePlanModifier struct{}

func (m hashedFilePlanModifier) Description(_ context.Context) string {
	return "Forces replacement when the attachment file's contents change (same contents under a different path do not force replacement)."
}

func (m hashedFilePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m hashedFilePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	planModifyHashedAttribute(req, resp, func(configRaw, stateRaw string) (bool, diag.Diagnostics) {
		configValue := HashedFileValue{StringValue: types.StringValue(configRaw)}
		stateValue := HashedFileValue{StringValue: types.StringValue(stateRaw)}
		return configValue.StringSemanticEquals(ctx, stateValue)
	})
}

type hashedContentPlanModifier struct{}

func (m hashedContentPlanModifier) Description(_ context.Context) string {
	return "Forces replacement when the attachment content changes."
}

func (m hashedContentPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m hashedContentPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	planModifyHashedAttribute(req, resp, func(configRaw, stateRaw string) (bool, diag.Diagnostics) {
		configValue := HashedContentValue{StringValue: types.StringValue(configRaw)}
		stateValue := HashedContentValue{StringValue: types.StringValue(stateRaw)}
		return configValue.StringSemanticEquals(ctx, stateValue)
	})
}
