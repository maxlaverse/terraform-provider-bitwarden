package bw

import (
	"context"
	"testing"

	test_command "github.com/maxlaverse/terraform-provider-bitwarden/internal/command/test"
	"github.com/stretchr/testify/assert"
)

func TestCreateObjectEncoding(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"create item eyJncm91cHMiOm51bGwsImxvZ2luIjp7fSwib2JqZWN0IjoiaXRlbSIsInNlY3VyZU5vdGUiOnt9LCJ0eXBlIjoxLCJmaWVsZHMiOlt7Im5hbWUiOiJ0ZXN0IiwidmFsdWUiOiJwYXNzZWQiLCJ0eXBlIjowLCJsaW5rZWRJZCI6bnVsbH1dfQ": `{}`,
	})
	defer removeMocks(t)

	b := NewClient("dummy")
	_, err := b.CreateObject(context.Background(), Object{
		Type:   ItemTypeLogin,
		Object: ObjectTypeItem,
		Fields: []Field{
			{
				Name:  "test",
				Value: "passed",
				Type:  0,
			},
		},
	})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "create item eyJncm91cHMiOm51bGwsImxvZ2luIjp7fSwib2JqZWN0IjoiaXRlbSIsInNlY3VyZU5vdGUiOnt9LCJ0eXBlIjoxLCJmaWVsZHMiOlt7Im5hbWUiOiJ0ZXN0IiwidmFsdWUiOiJwYXNzZWQiLCJ0eXBlIjowLCJsaW5rZWRJZCI6bnVsbH1dfQ", commandsExecuted()[0])
	}
}

func TestListObjects(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"list item --folderid folder-id --collectionid collection-id --search search": `[]`,
	})
	defer removeMocks(t)

	b := NewClient("dummy")
	_, err := b.ListObjects(context.Background(), "item", WithFolderID("folder-id"), WithCollectionID("collection-id"), WithSearch("search"))

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "list item --folderid folder-id --collectionid collection-id --search search", commandsExecuted()[0])
	}
}

func TestGetItem(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"get item object-id": `{}`,
	})
	defer removeMocks(t)

	b := NewClient("dummy")
	_, err := b.GetObject(context.Background(), Object{ID: "object-id", Object: ObjectTypeItem, Type: ItemTypeLogin})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "get item object-id", commandsExecuted()[0])
	}
}

func TestGetOrgCollection(t *testing.T) {
	removeMocks, commandsExecuted := test_command.MockCommands(t, map[string]string{
		"get org-collection object-id --organizationid org-id": `{}`,
	})
	defer removeMocks(t)

	b := NewClient("dummy")
	_, err := b.GetObject(context.Background(), Object{ID: "object-id", Object: ObjectTypeOrgCollection, OrganizationID: "org-id"})

	assert.NoError(t, err)
	if assert.Len(t, commandsExecuted(), 1) {
		assert.Equal(t, "get org-collection object-id --organizationid org-id", commandsExecuted()[0])
	}
}

func TestErrorContainsCommand(t *testing.T) {
	removeMocks, _ := test_command.MockCommands(t, map[string]string{
		"list org-collection --search search": ``,
	})
	defer removeMocks(t)

	b := NewClient("dummy")
	_, err := b.ListObjects(context.Background(), "org-collection", WithSearch("search"))

	if assert.Error(t, err) {
		assert.ErrorContains(t, err, "unable to parse result of 'list org-collection', error: 'unexpected end of JSON input', output: ''")
	}
}
