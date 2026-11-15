//go:build integration || integrationBws

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

func TestMain(m *testing.M) {
	if !IsOfficialBackend() {
		os.Exit(m.Run())
	}

	ctx := context.Background()
	sweepOfficialAccount(ctx, "before tests")
	code := m.Run()
	sweepOfficialAccount(ctx, "after tests")
	os.Exit(code)
}

func sweepOfficialAccount(ctx context.Context, when string) {
	if err := sweepOfficialPasswordManager(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: password manager sweep %s: %v\n", when, err)
	}
	if err := sweepOfficialSecretsManager(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: secrets manager sweep %s: %v\n", when, err)
	}
}

func sweepOfficialPasswordManager(ctx context.Context) error {
	acc := testConfiguration.Accounts[testAccountFullAdmin]
	if acc.ClientID == "" || acc.ClientSecret == "" || testConfiguration.ServerURL == "" {
		return nil
	}

	client := embedded.NewPasswordManagerClient(
		testConfiguration.ServerURL,
		testDeviceIdentifer,
		testDeviceVersion,
		embedded.DisableSyncAfterWrite(),
		embedded.DisableFailOnSyncAfterWriteVerification(),
	)
	if err := client.LoginWithAPIKey(ctx, acc.Password, acc.ClientID, acc.ClientSecret); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	if err := client.Sync(ctx); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	items, folders, collections := client.ListSyncedObjects()

	var errs []error
	for _, item := range items {
		if item.DeletedDate != nil || isProtectedID(item.ID) || !shouldSweepName(item.Name) {
			continue
		}
		if err := client.DeleteItem(ctx, models.Item{ID: item.ID, Object: models.ObjectTypeItem}); err != nil {
			errs = append(errs, fmt.Errorf("delete item %q (%s): %w", item.Name, item.ID, err))
			continue
		}
		fmt.Fprintf(os.Stderr, "swept item %q (%s)\n", item.Name, maskIdentifier(item.ID))
	}

	for _, folder := range folders {
		if isProtectedID(folder.ID) || !shouldSweepName(folder.Name) {
			continue
		}
		if err := client.DeleteFolder(ctx, models.Folder{ID: folder.ID, Object: models.ObjectTypeFolder}); err != nil {
			errs = append(errs, fmt.Errorf("delete folder %q (%s): %w", folder.Name, folder.ID, err))
			continue
		}
		fmt.Fprintf(os.Stderr, "swept folder %q (%s)\n", folder.Name, maskIdentifier(folder.ID))
	}

	for _, collection := range collections {
		if isProtectedID(collection.ID) || !shouldSweepName(collection.Name) {
			continue
		}
		if err := deleteCollectionAndItems(ctx, client, collection, items); err != nil {
			errs = append(errs, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "swept collection %q (%s)\n", collection.Name, maskIdentifier(collection.ID))
	}

	return joinSweepErrors(errs)
}

func sweepOfficialSecretsManager(ctx context.Context) error {
	token := os.Getenv("TEST_SECRETS_MANAGER_ACCESS_TOKEN")
	if token == "" || testConfiguration.ServerURL == "" {
		return nil
	}

	client := embedded.NewSecretsManagerClient(testConfiguration.ServerURL, testDeviceIdentifer, testDeviceVersion)
	if err := client.LoginWithAccessToken(ctx, token); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	secrets, err := client.ListSecrets(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for _, secret := range secrets {
		if !shouldSweepName(secret.Key) {
			continue
		}
		if err := client.DeleteSecret(ctx, models.Secret{ID: secret.ID}); err != nil {
			errs = append(errs, fmt.Errorf("delete secret %q (%s): %w", secret.Key, secret.ID, err))
			continue
		}
		fmt.Fprintf(os.Stderr, "swept secret %q (%s)\n", secret.Key, maskIdentifier(secret.ID))
	}

	projects, err := client.ListProjects(ctx)
	if err != nil {
		return err
	}
	for _, project := range projects {
		if !shouldSweepName(project.Name) {
			continue
		}
		if err := client.DeleteProject(ctx, models.Project{ID: project.ID}); err != nil {
			errs = append(errs, fmt.Errorf("delete project %q (%s): %w", project.Name, project.ID, err))
			continue
		}
		fmt.Fprintf(os.Stderr, "swept project %q (%s)\n", project.Name, maskIdentifier(project.ID))
	}

	return joinSweepErrors(errs)
}

func deleteCollectionAndItems(ctx context.Context, client embedded.PasswordManagerClient, collection models.OrgCollection, items []models.Item) error {
	for _, item := range items {
		if item.DeletedDate != nil || !itemInCollection(item, collection.ID) {
			continue
		}
		if err := client.DeleteItem(ctx, models.Item{ID: item.ID, Object: models.ObjectTypeItem}); err != nil {
			fmt.Fprintf(os.Stderr, "warning: delete item %q in collection %q: %v\n", item.Name, collection.Name, err)
		}
	}
	if err := client.DeleteOrganizationCollection(ctx, models.OrgCollection{
		ID:             collection.ID,
		OrganizationID: collection.OrganizationID,
		Object:         models.ObjectTypeOrgCollection,
	}); err != nil {
		return fmt.Errorf("delete collection %q (%s): %w", collection.Name, collection.ID, err)
	}
	return nil
}

func itemInCollection(item models.Item, collectionID string) bool {
	for _, id := range item.CollectionIds {
		if id == collectionID {
			return true
		}
	}
	return false
}

func isProtectedID(id string) bool {
	if id == "" {
		return false
	}
	return id == testConfiguration.Resources.CollectionID ||
		id == testConfiguration.Resources.FolderID ||
		id == testConfiguration.Resources.OrganizationID ||
		id == testConfiguration.Resources.GroupID
}

func shouldSweepName(name string) bool {
	if strings.HasPrefix(name, testResourceNamePrefix+"-") {
		return true
	}
	switch name {
	case "login-bar", "secure-bar", "ssh-key-bar", "folder-bar", "folder-new-name-bar",
		"org-col-bar", "org-col-new-name-bar", "foo", "project-foo", "project-bar", "test":
		return true
	}
	if strings.HasPrefix(name, "Login Item ") || strings.HasPrefix(name, "login-test-col-") {
		return true
	}
	if strings.HasPrefix(name, "another item with username") {
		return true
	}
	return false
}

func joinSweepErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	parts := make([]string, 0, len(errs))
	for _, err := range errs {
		parts = append(parts, err.Error())
	}
	return fmt.Errorf("%s", strings.Join(parts, "; "))
}
