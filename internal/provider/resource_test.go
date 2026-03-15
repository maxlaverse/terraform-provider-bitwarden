package provider

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/schema_definition"
)

const (
	regExpId   = `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`
	regExpDate = `^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z)?$`
)

func checkObject(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeName, regexp.MustCompile("^([a-z0-9-]+)-bar$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeID, regexp.MustCompile(regExpId),
		),
	)
}

func checkItemBase(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		checkObject(resourceName),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeNotes, regexp.MustCompile("^notes([A-Za-z-]*)$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeFolderID, regexp.MustCompile(regExpId),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeReprompt, regexp.MustCompile("^true"),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeRevisionDate, regexp.MustCompile(regExpDate),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeCreationDate, regexp.MustCompile(regExpDate),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeDeletedDate, regexp.MustCompile(regExpDate),
		),
		resource.TestMatchResourceAttr(
			resourceName, schema_definition.AttributeOrganizationID, regexp.MustCompile(regExpId),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.#", schema_definition.AttributeCollectionIDs), regexp.MustCompile("^1$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.0", schema_definition.AttributeCollectionIDs), regexp.MustCompile(regExpId),
		),
		checkItemFields(resourceName),
	)
}

func checkItemFields(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.#", schema_definition.AttributeField), regexp.MustCompile("^3$"),
		),
		resource.TestCheckResourceAttr(
			resourceName, fmt.Sprintf("%s.0.text", schema_definition.AttributeField), "value-text",
		),
		resource.TestCheckResourceAttr(
			resourceName, fmt.Sprintf("%s.1.boolean", schema_definition.AttributeField), "true",
		),
		resource.TestCheckResourceAttr(
			resourceName, fmt.Sprintf("%s.2.hidden", schema_definition.AttributeField), "value-hidden",
		),
	)
}

// checkResourceAttrListEqualsSet verifies that the resource's list attribute contains exactly the expected string values, in any order.
func checkResourceAttrListEqualsSet(resourceName, attr string, expectedValues []*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		countStr, ok := rs.Primary.Attributes[attr+".#"]
		if !ok {
			return fmt.Errorf("%s.# not found", attr)
		}
		var count int
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return fmt.Errorf("invalid %s.# value: %s", attr, countStr)
		}
		actual := make([]string, 0, count)
		for i := 0; i < count; i++ {
			v := rs.Primary.Attributes[fmt.Sprintf("%s.%d", attr, i)]
			actual = append(actual, v)
		}
		if len(actual) != len(expectedValues) {
			return fmt.Errorf("%s length: got %d, want %d (got %v, want %v)", attr, len(actual), len(expectedValues), actual, expectedValues)
		}
		for _, expected := range expectedValues {
			if expected == nil {
				return fmt.Errorf("%s missing expected value (got %v)", attr, actual)
			}
			if !slices.Contains(actual, *expected) {
				return fmt.Errorf("%s missing expected value %q (got %v)", attr, *expected, actual)
			}
		}
		return nil
	}
}
