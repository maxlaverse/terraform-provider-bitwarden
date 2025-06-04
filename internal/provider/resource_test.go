package provider

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func conditionalAssertion(shouldRun bool, testCheckFunc ...resource.TestCheckFunc) resource.TestCheckFunc {
	if !shouldRun {
		return resource.ComposeTestCheckFunc()
	}
	return resource.ComposeTestCheckFunc(testCheckFunc...)
}
