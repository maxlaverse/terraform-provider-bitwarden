package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/embedded"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

func ensureTestResourcesExist(t *testing.T) {
	if testConfiguration.WasResourcesCreationAttempted(t) {
		return
	}

	if IsOfficialBackend() {
		fmt.Fprint(os.Stderr, testConfiguration.PrintConfiguration())
		return
	}

	ctx := t.Context()
	bwClient := bwEmbeddedTestClient(t).(embedded.PasswordManagerClient)

	testOrgId, err := bwClient.CreateOrganization(ctx, "org-"+testConfiguration.UniqueTestIdentifier, "coll-"+testConfiguration.UniqueTestIdentifier, testConfiguration.Accounts[testAccountFullAdmin].Email)
	if err != nil {
		t.Fatal(err)
	}
	testConfiguration.Resources.OrganizationID = testOrgId
	t.Logf("Created organization %s", testConfiguration.Resources.OrganizationID)

	for _, account := range []testAccountName{testAccountOrgOwner, testAccountOrgUser, testAccountOrgAdmin, testAccountOrgManager} {
		err = bwClient.InviteUser(ctx, testConfiguration.Resources.OrganizationID, testConfiguration.Accounts[account].Email, testConfiguration.Accounts[account].RoleInTestOrganization)
		if err != nil {
			t.Fatal(err)
		}

		userIdInOrg, err := bwClient.ConfirmInvite(ctx, testConfiguration.Resources.OrganizationID, testConfiguration.Accounts[account].Email)
		if err != nil {
			t.Fatal(err)
		}
		ac := testConfiguration.Accounts[account]
		ac.UserIdInTestOrganization = userIdInOrg
		testConfiguration.Accounts[account] = ac
		t.Logf("Invited %s to organization %s (%s)", testConfiguration.Accounts[account].Email, testConfiguration.Resources.OrganizationID, testConfiguration.Accounts[account].UserIdInTestOrganization)
	}

	webapiClient := webapi.NewClient(testConfiguration.ServerURL, embedded.NewDeviceIdentifier(), testDeviceVersion)
	_, err = webapiClient.LoginWithAPIKey(ctx, testConfiguration.Accounts[testAccountFullAdmin].ClientID, testConfiguration.Accounts[testAccountFullAdmin].ClientSecret)
	if err != nil {
		t.Fatal(err)
	}
	cols, err := webapiClient.GetOrganizationCollections(ctx, testConfiguration.Resources.OrganizationID)
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) == 0 {
		t.Fatal("no collections found")
	}
	testConfiguration.Resources.CollectionID = cols[0].Id

	testFolderName := fmt.Sprintf("folder-%s-bar", testConfiguration.UniqueTestIdentifier)
	folder, err := bwClient.CreateFolder(ctx, models.Folder{
		Object: models.ObjectTypeFolder,
		Name:   testFolderName,
	})
	if err != nil {
		t.Fatal(err)
	}

	testConfiguration.Resources.FolderID = folder.ID
	t.Logf("Created Folder '%s' (%s)", testFolderName, testConfiguration.Resources.FolderID)

	testConfiguration.Resources.GroupName = fmt.Sprintf("group-%s-bar", testConfiguration.UniqueTestIdentifier)
	group, err := bwClient.CreateOrganizationGroup(ctx, models.OrgGroup{
		OrganizationID: testConfiguration.Resources.OrganizationID,
		Name:           testConfiguration.Resources.GroupName,
		Collections:    []models.OrgCollectionMember{},
		Users:          []models.OrgCollectionMember{},
	})
	if err != nil {
		t.Fatal(err)
	}

	testConfiguration.Resources.GroupID = group.ID
	t.Logf("Created Group '%s' (%s)", testConfiguration.Resources.GroupName, testConfiguration.Resources.GroupID)

	err = bwClient.Sync(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Synced embedded test client")
	fmt.Fprint(os.Stderr, testConfiguration.PrintConfiguration())
}

func ensureTestAccountsExist(t *testing.T) {
	if testConfiguration.WasAccountCreationAttempted(t) {
		return
	}

	clearTestVault(t)

	client := embedded.NewPasswordManagerClient(testConfiguration.ServerURL, testDeviceIdentifer, testDeviceVersion)

	kdfConfig := models.KdfConfiguration{
		KdfType:        models.KdfTypePBKDF2_SHA256,
		KdfIterations:  testKdfIterations,
		KdfMemory:      0,
		KdfParallelism: 0,
	}

	err := client.RegisterUser(t.Context(), testConfiguration.Accounts[testAccountFullAdmin].Name, testConfiguration.Accounts[testAccountFullAdmin].Email, testConfiguration.Accounts[testAccountFullAdmin].Password, kdfConfig)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
		t.Fatal(err)
	}
	t.Logf("Registered test user (%s) %s", testAccountFullAdmin, testConfiguration.Accounts[testAccountFullAdmin].Email)

	err = client.LoginWithPassword(t.Context(), testConfiguration.Accounts[testAccountFullAdmin].Email, testConfiguration.Accounts[testAccountFullAdmin].Password)
	if err != nil {
		t.Fatal(err)
	}

	apiKey, err := client.GetAPIKey(t.Context(), testConfiguration.Accounts[testAccountFullAdmin].Email, testConfiguration.Accounts[testAccountFullAdmin].Password)
	if err != nil {
		t.Fatal(err)
	}
	acc := testConfiguration.Accounts[testAccountFullAdmin]
	acc.ClientID = apiKey.ClientID
	acc.ClientSecret = apiKey.ClientSecret
	testConfiguration.Accounts[testAccountFullAdmin] = acc

	client.Logout(t.Context())

	kdfConfig = models.KdfConfiguration{
		KdfType:        models.KdfTypeArgon2,
		KdfIterations:  3,
		KdfMemory:      64,
		KdfParallelism: 4,
	}

	for _, account := range []testAccountName{testAccountOrgOwner, testAccountOrgUser, testAccountOrgAdmin, testAccountOrgManager} {
		err = client.RegisterUser(t.Context(), testConfiguration.Accounts[account].Name, testConfiguration.Accounts[account].Email, testConfiguration.Accounts[account].Password, kdfConfig)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "user already exists") {
			t.Fatal(err)
		}
		t.Logf("Registered test user (%s) %s", account, testConfiguration.Accounts[account].Email)

		err = client.LoginWithPassword(t.Context(), testConfiguration.Accounts[account].Email, testConfiguration.Accounts[account].Password)
		if err != nil {
			t.Fatal(err)
		}

		apiKey, err := client.GetAPIKey(t.Context(), testConfiguration.Accounts[account].Email, testConfiguration.Accounts[account].Password)
		if err != nil {
			t.Fatal(err)
		}
		acc := testConfiguration.Accounts[account]
		acc.ClientID = apiKey.ClientID
		acc.ClientSecret = apiKey.ClientSecret
		testConfiguration.Accounts[account] = acc

		client.Logout(t.Context())
	}
	testConfiguration.didAccountCreationSucceed = true
}

func clearTestVault(t *testing.T) {
	err := os.Remove(".bitwarden/data.json")
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
