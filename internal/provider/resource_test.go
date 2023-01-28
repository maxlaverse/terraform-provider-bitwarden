package provider

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	regExpId = "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"
)

func checkObject(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			resourceName, attributeName, regexp.MustCompile("^([a-z0-9]+)-bar$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, attributeID, regexp.MustCompile(regExpId),
		),
	)
}

func checkItemGeneral(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		checkObject(resourceName),
		resource.TestMatchResourceAttr(
			resourceName, attributeNotes, regexp.MustCompile("^notes$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, attributeFolderID, regexp.MustCompile(regExpId),
		),
		resource.TestMatchResourceAttr(
			resourceName, attributeReprompt, regexp.MustCompile("^true"),
		),
		resource.TestMatchResourceAttr(
			resourceName, attributeFavorite, regexp.MustCompile("^true"),
		),
		resource.TestMatchResourceAttr(
			resourceName, attributeOrganizationID, regexp.MustCompile(regExpId),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.#", attributeCollectionIDs), regexp.MustCompile("^1$"),
		),
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.0", attributeCollectionIDs), regexp.MustCompile(regExpId),
		),
		checkItemFields(resourceName),
	)
}

func checkItemFields(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestMatchResourceAttr(
			resourceName, fmt.Sprintf("%s.#", attributeField), regexp.MustCompile("^3$"),
		),
		resource.TestCheckResourceAttr(
			resourceName, fmt.Sprintf("%s.0.text", attributeField), "value-text",
		),
		resource.TestCheckResourceAttr(
			resourceName, fmt.Sprintf("%s.1.boolean", attributeField), "true",
		),
		resource.TestCheckResourceAttr(
			resourceName, fmt.Sprintf("%s.2.hidden", attributeField), "value-hidden",
		),
	)
}
