package bwcli

import (
	"context"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	test_command "github.com/maxlaverse/terraform-provider-bitwarden/internal/command/test"
	"github.com/stretchr/testify/assert"
)

func TestCreateObjectEncoding(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"create item eyJmaWVsZHMiOlt7Im5hbWUiOiJ0ZXN0IiwidmFsdWUiOiJwYXNzZWQiLCJ0eXBlIjowLCJsaW5rZWRJZCI6bnVsbH1dLCJsb2dpbiI6e30sIm9iamVjdCI6Iml0ZW0iLCJzZWN1cmVOb3RlIjp7fSwidHlwZSI6MX0": `{}`,
	})
	defer removeMocks(t)

	b := NewPasswordManagerClient("dummy")
	_, err := b.CreateItem(context.Background(), models.Item{
		Object: models.ObjectTypeItem,
		Type:   models.ItemTypeLogin,
		Fields: []models.Field{
			{
				Name:  "test",
				Value: "passed",
				Type:  0,
			},
		},
	})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "create item eyJmaWVsZHMiOlt7Im5hbWUiOiJ0ZXN0IiwidmFsdWUiOiJwYXNzZWQiLCJ0eXBlIjowLCJsaW5rZWRJZCI6bnVsbH1dLCJsb2dpbiI6e30sIm9iamVjdCI6Iml0ZW0iLCJzZWN1cmVOb3RlIjp7fSwidHlwZSI6MX0", commandsExecuted()[0])
	}
}

func TestCreateOrgCollection(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"create org-collection eyJuYW1lIjoidGVzdCIsIm9iamVjdCI6Im9yZy1jb2xsZWN0aW9uIiwib3JnYW5pemF0aW9uSWQiOiJvcmctaWQiLCJncm91cHMiOltdfQ --organizationid org-id": `{}`,
	})
	defer removeMocks(t)

	b := NewPasswordManagerClient("dummy")
	_, err := b.CreateOrganizationCollection(context.Background(), models.OrgCollection{
		Object:         models.ObjectTypeOrgCollection,
		Name:           "test",
		OrganizationID: "org-id",
	})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "create org-collection eyJuYW1lIjoidGVzdCIsIm9iamVjdCI6Im9yZy1jb2xsZWN0aW9uIiwib3JnYW5pemF0aW9uSWQiOiJvcmctaWQiLCJncm91cHMiOltdfQ --organizationid org-id", commandsExecuted()[0])
	}
}

func TestEditOrgCollection(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"edit org-collection 1234 --organizationid org-id eyJpZCI6IjEyMzQiLCJuYW1lIjoidGVzdCIsIm9iamVjdCI6Im9yZy1jb2xsZWN0aW9uIiwib3JnYW5pemF0aW9uSWQiOiJvcmctaWQiLCJncm91cHMiOltdfQ": `{}`,
		"sync": `{}`,
	})
	defer removeMocks(t)

	b := NewPasswordManagerClient("dummy")
	_, err := b.EditOrganizationCollection(context.Background(), models.OrgCollection{
		Object:         models.ObjectTypeOrgCollection,
		ID:             "1234",
		Name:           "test",
		OrganizationID: "org-id",
	})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 2) {
		assert.Equal(t, "edit org-collection 1234 --organizationid org-id eyJpZCI6IjEyMzQiLCJuYW1lIjoidGVzdCIsIm9iamVjdCI6Im9yZy1jb2xsZWN0aW9uIiwib3JnYW5pemF0aW9uSWQiOiJvcmctaWQiLCJncm91cHMiOltdfQ", commandsExecuted()[0])
	}
}

func TestDeleteOrgCollection(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"delete org-collection 1234 --organizationid org-id": `{}`,
	})
	defer removeMocks(t)

	b := NewPasswordManagerClient("dummy")
	err := b.DeleteOrganizationCollection(context.Background(), models.OrgCollection{
		Object:         models.ObjectTypeOrgCollection,
		ID:             "1234",
		Name:           "test",
		OrganizationID: "org-id",
	})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "delete org-collection 1234 --organizationid org-id", commandsExecuted()[0])
	}
}

func TestListObjects(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"list items --folderid folder-id --collectionid collection-id --search search": `[{ "id": "object-id" }]`,
	})
	defer removeMocks(t)

	b := NewPasswordManagerClient("dummy")
	_, err := b.FindItem(context.Background(), bitwarden.WithFolderID("folder-id"), bitwarden.WithCollectionID("collection-id"), bitwarden.WithSearch("search"))

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "list items --folderid folder-id --collectionid collection-id --search search", commandsExecuted()[0])
	}
}

func TestGetItem(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"get item object-id": `{}`,
	})
	defer removeMocks(t)

	b := NewPasswordManagerClient("dummy")
	_, err := b.GetItem(context.Background(), models.Item{ID: "object-id", Object: models.ObjectTypeItem, Type: models.ItemTypeLogin})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "get item object-id", commandsExecuted()[0])
	}
}

func TestGetOrganizationCollection(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"get org-collection object-id --organizationid org-id": `{}`,
	})
	defer removeMocks(t)

	b := NewPasswordManagerClient("dummy")
	_, err := b.GetOrganizationCollection(context.Background(), models.OrgCollection{ID: "object-id", Object: models.ObjectTypeOrgCollection, OrganizationID: "org-id"})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "get org-collection object-id --organizationid org-id", commandsExecuted()[0])
	}
}

func TestErrorContainsCommand(t *testing.T) {
	removeMocks, _ := test_command.MockCommands(t, map[string]string{
		"list org-collections --search search": ``,
	})
	defer removeMocks(t)

	b := NewPasswordManagerClient("dummy")
	_, err := b.FindOrganizationCollection(context.Background(), bitwarden.WithSearch("search"))

	if assert.Error(t, err) {
		assert.ErrorContains(t, err, "unable to parse result of 'list org-collections', error: 'unexpected end of JSON input', output: ''")
	}
}
