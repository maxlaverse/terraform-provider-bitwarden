package embedded

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/encryptedstring"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

type PasswordManagerClient interface {
	BaseVault
	ConfirmInvite(ctx context.Context, orgId, userEmail string) (string, error)
	CreateFolder(ctx context.Context, obj models.Folder) (*models.Folder, error)
	CreateGroup(ctx context.Context, obj models.Group) (*models.Group, error)
	CreateItem(ctx context.Context, obj models.Item) (*models.Item, error)
	CreateOrganization(ctx context.Context, organizationName, organizationLabel, billingEmail string) (string, error)
	CreateOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)
	CreateAttachmentFromContent(ctx context.Context, itemId, filename string, content []byte) (*models.Item, error)
	CreateAttachmentFromFile(ctx context.Context, itemId, filePath string) (*models.Item, error)
	DeleteAttachment(ctx context.Context, itemId, attachmentId string) error
	DeleteFolder(ctx context.Context, obj models.Folder) error
	DeleteGroup(ctx context.Context, obj models.Group) error
	DeleteItem(ctx context.Context, obj models.Item) error
	DeleteOrganizationCollection(ctx context.Context, obj models.OrgCollection) error
	EditFolder(ctx context.Context, obj models.Folder) (*models.Folder, error)
	EditGroup(ctx context.Context, obj models.Group) (*models.Group, error)
	EditItem(ctx context.Context, obj models.Item) (*models.Item, error)
	EditOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)
	FindOrganizationMember(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgMember, error)
	FindOrganizationCollection(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgCollection, error)
	GetAPIKey(ctx context.Context, username, password string) (*models.ApiKey, error)
	GetAttachment(ctx context.Context, itemId, attachmentId string) ([]byte, error)
	GetGroup(ctx context.Context, obj models.Group) (*models.Group, error)
	GetOrganization(context.Context, models.Organization) (*models.Organization, error)
	GetOrganizationMember(context.Context, models.OrgMember) (*models.OrgMember, error)
	GetOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)
	InviteUser(ctx context.Context, orgId, userEmail string, memberRoleType models.OrgMemberRoleType) error
	IsSyncAfterWriteVerificationDisabled() bool
	LoginWithAPIKey(ctx context.Context, password, clientId, clientSecret string) error
	LoginWithPassword(ctx context.Context, username, password string) error
	Logout(ctx context.Context) error
	RegisterUser(ctx context.Context, name, username, password string, kdfConfig models.KdfConfiguration) error
	Sync(ctx context.Context) error
	Unlock(ctx context.Context, password string) error
}

type PasswordManagerOptions func(c bitwarden.PasswordManager)

// DisableCryptoSafeMode disables the safe mode for crypto operations, which reverses
// crypto.Encrypt() to make sure it can decrypt the result.
func DisableCryptoSafeMode() PasswordManagerOptions {
	return func(c bitwarden.PasswordManager) {
		crypto.SafeMode = false
	}
}

// DisableObjectEncryptionVerification disables the systematic attempts to decrypt objects
// (items, folders, collections) after they have been created or edited, to verify that the
// encryption can be reverse.
func DisableObjectEncryptionVerification() PasswordManagerOptions {
	return func(c bitwarden.PasswordManager) {
		c.(*webAPIVault).baseVault.verifyObjectEncryption = false
	}
}

// DisableSyncAfterWrite disables the systematic Sync() after a write operation (create, edit,
// delete) to the vault. Write operations already return the object that was created or edited, so
// Sync() is not strictly necessary.
func DisableSyncAfterWrite() PasswordManagerOptions {
	return func(c bitwarden.PasswordManager) {
		c.(*webAPIVault).syncAfterWrite = false
	}
}

// DisableFailOnSyncAfterWriteVerification disables the check for differences between the local and
// remote objects after a write operation (create, edit, delete).
func DisableFailOnSyncAfterWriteVerification() PasswordManagerOptions {
	return func(c bitwarden.PasswordManager) {
		c.(*webAPIVault).failOnSyncAfterWriteVerification = false
	}
}

func WithPasswordManagerHttpOptions(opts ...webapi.Options) PasswordManagerOptions {
	return func(c bitwarden.PasswordManager) {
		c.(*webAPIVault).clientOpts = opts
	}
}

// Panic on error is useful for debugging, but should not be used in production.
func EnablePanicOnEncryptionError() PasswordManagerOptions {
	return func(c bitwarden.PasswordManager) {
		panicOnEncryptionErrors = true
	}
}

func NewPasswordManagerClient(serverURL, deviceIdentifier, providerVersion string, opts ...PasswordManagerOptions) PasswordManagerClient {
	c := &webAPIVault{
		baseVault: baseVault{
			collectionDetailsLoadedForOrg: map[string]bool{},
			objectStore:                   make(map[string]interface{}),
			verifyObjectEncryption:        true,
		},
		serverURL: serverURL,

		// Always run Sync() after creating, editing, or deleting an object and verify the result
		// by comparing the local and remote objects.
		syncAfterWrite:                   true,
		failOnSyncAfterWriteVerification: true,
	}

	for _, o := range opts {
		o(c)
	}

	c.client = webapi.NewClient(serverURL, deviceIdentifier, providerVersion, c.clientOpts...)

	return c
}

func NewDeviceIdentifier() string {
	return uuid.New().String()
}

type webAPIVault struct {
	baseVault
	client     webapi.Client
	clientOpts []webapi.Options

	syncAfterWrite                   bool
	failOnSyncAfterWriteVerification bool
	serverURL                        string
}

func (v *webAPIVault) ConfirmInvite(ctx context.Context, orgId, userEmail string) (string, error) {
	// Write lock is needed since we eventually load the organization members.
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	err := v.ensureUsersLoadedForOrg(ctx, orgId)
	if err != nil {
		return "", fmt.Errorf("error loading users of organization '%s': %w", orgId, err)
	}

	orgUser, err := v.organizationMembers.FindMemberByEmail(orgId, userEmail)
	if err != nil {
		return "", fmt.Errorf("error getting organization user : %w", err)
	}

	publicKey, err := v.getUserPublicKey(ctx, orgUser.UserId)
	if err != nil {
		return "", fmt.Errorf("error getting user public key: %w", err)
	}

	orgKey, err := keybuilder.RSAEncrypt(v.loginAccount.Secrets.OrganizationSecrets[orgId].Key.Key, publicKey)
	if err != nil {
		return "", fmt.Errorf("error rsa encrypting organization key: %w", err)
	}

	return orgUser.ID, v.client.ConfirmOrganizationUser(ctx, orgId, orgUser.ID, string(orgKey))
}

func (v *webAPIVault) CreateAttachmentFromContent(ctx context.Context, itemId, filename string, content []byte) (*models.Item, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	return v.createAttachment(ctx, itemId, filename, content)
}

func (v *webAPIVault) CreateAttachmentFromFile(ctx context.Context, itemId, filePath string) (*models.Item, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	filename := filepath.Base(filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading attachment file: %w", err)
	}

	return v.createAttachment(ctx, itemId, filename, data)
}

func (v *webAPIVault) createAttachment(ctx context.Context, itemId, filename string, content []byte) (*models.Item, error) {
	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	req, data, err := v.prepareAttachmentCreationRequest(ctx, itemId, filename, content)
	if err != nil {
		return nil, fmt.Errorf("error preparing attachment creation request: %w", err)
	}

	resp, err := v.client.CreateObjectAttachment(ctx, itemId, data, *req)
	if err != nil {
		return nil, fmt.Errorf("error creating attachment: %w", err)
	}

	switch resp.FileUploadType {
	case models.FileUploadTypeDirect:
		err = v.client.CreateObjectAttachmentData(ctx, itemId, resp.AttachmentId, data)
		if err != nil {
			return nil, fmt.Errorf("error creating attachment data: %w", err)
		}
	case models.FileUploadTypeAzure:
		err = v.client.UploadContentToUrl(ctx, webapi.CloudStorageProviderAzure, resp.Url, data)
		if err != nil {
			return nil, fmt.Errorf("error uploading data to Azure: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file upload type: %d", resp.FileUploadType)
	}

	resObj, err := decryptItem((*resp).CipherResponse, v.loginAccount.Secrets)
	if err != nil {
		return nil, fmt.Errorf("error decrypting resulting obj data attachment: %w", err)
	}

	v.storeObject(ctx, *resObj)

	if v.syncAfterWrite {
		err = v.sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := getObject(v.objectStore, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting object after attachment upload (sync-after-write): %w", err)
		}

		// The attachment's URL contains a signed token generated on each request. We need to diff
		// it out if we want the comparison to work.
		return remoteObj, v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj, "/attachments/*/url", "/revisionDate")
	}
	return resObj, nil
}

func (v *webAPIVault) CreateFolder(ctx context.Context, obj models.Folder) (*models.Folder, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	var resObj *models.Folder

	encObj, err := encryptFolder(ctx, obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
	if err != nil {
		return nil, fmt.Errorf("error encrypting folder for creation: %w", err)
	}

	resEncFolder, err := v.client.CreateFolder(ctx, *encObj)
	if err != nil {
		return nil, fmt.Errorf("error creating folder: %w", err)
	}

	resObj, err = decryptFolder(*resEncFolder, v.loginAccount.Secrets)
	if err != nil {
		return nil, fmt.Errorf("error decrypting folder after creation: %w", err)
	}

	v.storeObject(ctx, *resObj)

	if v.syncAfterWrite {
		err := v.sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := getObject(v.objectStore, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting folder after creation (sync-after-write): %w", err)
		}

		return remoteObj, v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj, "/revisionDate")
	}
	return resObj, nil
}

func (v *webAPIVault) CreateGroup(ctx context.Context, obj models.Group) (*models.Group, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	resObj, err := v.client.CreateGroup(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("error creating group: %w", err)
	}

	if resObj.Collections == nil {
		resObj.Collections = []models.OrgCollectionMember{}
	}

	if v.syncAfterWrite {
		remoteObj, err := v.getGroup(ctx, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting group after creation (sync-after-write): %w", err)
		}

		return remoteObj, v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj)
	}
	return resObj, nil
}

func (v *webAPIVault) CreateItem(ctx context.Context, obj models.Item) (*models.Item, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	var resObj *models.Item
	encObj, err := encryptItem(ctx, obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item for creation: %w", err)
	}

	resEncObj, err := v.client.CreateItem(ctx, *encObj)

	if err != nil {
		return nil, fmt.Errorf("error creating item: %w", err)
	}

	resEncObj.Object = obj.Object
	resEncObj.Type = obj.Type
	resObj, err = decryptItem(*resEncObj, v.loginAccount.Secrets)
	if err != nil {
		return nil, fmt.Errorf("error decrypting item after creation: %w", err)
	}

	v.storeObject(ctx, *resObj)

	if v.syncAfterWrite {
		err := v.sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := getObject(v.objectStore, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting item after creation (sync-after-write): %w", err)
		}

		// NOTE: The official Bitwarden server returns dates that are a few milliseconds apart
		//       between the object's creation call and a later retrieval. We need to ignore
		//       these differences in the diff.
		// NOTE: The official Bitwarden server don't return the collectionIds in the response
		//       for items, even if they're actually taken into account.
		return remoteObj, v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj, "/creationDate", "/revisionDate", "/collectionIds")
	}
	return resObj, nil
}

func (v *webAPIVault) ensureUsersLoadedForOrg(ctx context.Context, orgId string) error {
	if v.organizationMembers.OrganizationInitialized(orgId) {
		return nil
	}

	orgUsers, err := v.client.GetOrganizationUsers(ctx, orgId)
	if err != nil {
		return fmt.Errorf("error getting organization users: %w", err)
	}

	v.organizationMembers.LoadMembers(orgId, orgUsers)

	return nil
}

func (v *webAPIVault) CreateOrganizationCollection(ctx context.Context, obj models.OrgCollection) (*models.OrgCollection, error) {
	// ValidateFunc is not supported on TypeSet, which means we can't check for
	// duplicate during Schema validation. Doing it here instead.
	err := checkForDuplicateMembers(obj.Users)
	if err != nil {
		return nil, err
	}

	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	err = v.checkMembersExistence(ctx, obj.OrganizationID, obj.Users)
	if err != nil {
		return nil, err
	}

	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	manageMembership := len(obj.Users) > 0
	obj.Manage = manageMembership

	encObj, err := encryptOrgCollection(ctx, obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
	if err != nil {
		return nil, fmt.Errorf("error encrypting collection for creation: %w", err)
	}

	resOrgCol, err := v.client.CreateOrganizationCollection(ctx, obj.OrganizationID, *encObj)
	if err != nil {
		return nil, fmt.Errorf("error creating collection: %w", err)
	}

	resObj, err := decryptOrgCollection(*resOrgCol, v.loginAccount.Secrets)
	if err != nil {
		return nil, fmt.Errorf("error decrypting collection after creation: %w", err)
	}

	if manageMembership {
		err := v.reloadCollectionDetailsOfOrg(ctx, obj.OrganizationID)
		if err != nil {
			return nil, fmt.Errorf("error loading collections: %w", err)
		}

		resObj, err = getObject(v.objectStore, *resObj)
		if err != nil {
			return nil, fmt.Errorf("refetch-after-write error: %w", err)
		}
	} else {
		v.storeObject(ctx, *resObj)
	}

	if v.syncAfterWrite && !manageMembership {
		err := v.sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := getObject(v.objectStore, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting collection after creation (sync-after-write): %w", err)
		}

		return remoteObj, v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj)
	} else if v.syncAfterWrite && manageMembership {
		// If we had enough permissions to manage memberships, the server will
		// always return the collection with the Manage flag set to true.
		return resObj, v.verifyObjectAfterWrite(ctx, *resObj, obj, "/id", "/manage")
	}
	return resObj, err

}

func (v *webAPIVault) EditOrganizationCollection(ctx context.Context, obj models.OrgCollection) (*models.OrgCollection, error) {
	// ValidateFunc is not supported on TypeSet, which means we can't check for
	// duplicate during Schema validation. Doing it here instead.
	err := checkForDuplicateMembers(obj.Users)
	if err != nil {
		return nil, err
	}

	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	err = v.checkMembersExistence(ctx, obj.OrganizationID, obj.Users)
	if err != nil {
		return nil, err
	}

	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	// When editing a collection, we need to ensure we have enough permissions
	// to manage the collection's memberships if members were specified.
	currentObj, err := v.getOrganizationCollection(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("error getting collection prior to edition: %w %+v", err, obj)
	}

	manageMembership := currentObj.Manage
	if !manageMembership && len(obj.Users) > 0 {
		return nil, fmt.Errorf("error editing collection: you need to have the Manage permission to edit memberships")
	}

	encObj, err := encryptOrgCollection(ctx, obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
	if err != nil {
		return nil, fmt.Errorf("error encrypting collection for edition: %w", err)
	}

	resOrgCol, err := v.client.EditOrganizationCollection(ctx, obj.OrganizationID, obj.ID, *encObj)
	if err != nil {
		return nil, fmt.Errorf("error decrypting collection after creation: %w", err)
	}

	resObj, err := decryptOrgCollection(*resOrgCol, v.loginAccount.Secrets)
	if err != nil {
		return nil, fmt.Errorf("error decrypting collection after creation: %w", err)
	}

	if manageMembership {
		err := v.reloadCollectionDetailsOfOrg(ctx, obj.OrganizationID)
		if err != nil {
			return nil, fmt.Errorf("error loading collections: %w", err)
		}

		resObj, err = getObject(v.objectStore, *resObj)
		if err != nil {
			return nil, fmt.Errorf("refetch-after-write error: %w", err)
		}
	} else {
		v.storeObject(ctx, *resObj)
	}

	if v.syncAfterWrite && !manageMembership {
		err := v.sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := getObject(v.objectStore, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting collection after edition (sync-after-write): %w", err)
		}

		return remoteObj, v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj)
	} else if v.syncAfterWrite && manageMembership {
		// The server adapts the member list based on the permissions of the
		// users. Typically, if we set Manage=false but the user is an owner,
		// the server will return the collection with Manage=true and the
		// comparison fails. Similarly, if we add the owner of a collection as
		// user, it won't be returned in the response as it's an implicit
		// member. Instead of managing these edge cases, we'll just not compare
		// the objects in this case.
		return resObj, err

	}
	return resObj, err
}

func (v *webAPIVault) CreateOrganization(ctx context.Context, organizationName, organizationLabel, billingEmail string) (string, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return "", models.ErrVaultLocked
	}

	encSharedKey, sharedKey, err := keybuilder.GenerateSharedKey(&v.loginAccount.Secrets.RSAPrivateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("error generating shared key: %w", err)
	}

	collectionName, err := crypto.EncryptAsString([]byte(organizationLabel), *sharedKey)
	if err != nil {
		return "", fmt.Errorf("error encryption collection label: %w", err)
	}

	publicKey, encryptedPrivateKey, err := keybuilder.GenerateEncryptedRSAKeyPair(*sharedKey)
	if err != nil {
		return "", fmt.Errorf("error generating key pair: %w", err)
	}

	orgCreationRequest := webapi.CreateOrganizationRequest{
		Name:           organizationName,
		BillingEmail:   billingEmail,
		CollectionName: collectionName,
		Key:            encSharedKey,
		Keys: webapi.KeyPair{
			PublicKey:           publicKey,
			EncryptedPrivateKey: encryptedPrivateKey,
		},
		PlanType: 0,
	}
	res, err := v.client.CreateOrganization(ctx, orgCreationRequest)
	if err != nil {
		return "", fmt.Errorf("error creating organization: %w", err)
	}

	v.baseVault.loginAccount.Secrets.OrganizationSecrets[res.Id] = OrganizationSecret{
		OrganizationUUID: res.Id,
		Name:             organizationName,
		Key:              *sharedKey,
	}

	v.storeOrganizationSecrets(ctx)

	if v.syncAfterWrite {
		orgSecretBeforeSync := v.baseVault.loginAccount.Secrets.OrganizationSecrets[res.Id]
		err := v.sync(ctx)
		if err != nil {
			return "", fmt.Errorf("sync-after-write error: %w", err)
		}

		return res.Id, v.verifyObjectAfterWrite(ctx, orgSecretBeforeSync, v.baseVault.loginAccount.Secrets.OrganizationSecrets[res.Id])
	}

	return res.Id, nil
}

func (v *webAPIVault) DeleteAttachment(ctx context.Context, itemId, attachmentId string) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return models.ErrVaultLocked
	}

	// TODO: Don't fail if attachment is already gone
	err := v.client.DeleteObjectAttachment(ctx, itemId, attachmentId)
	if err != nil {
		return fmt.Errorf("error deleting attachment: %w", err)
	}

	resObj, err := getObject(v.objectStore, models.Item{ID: itemId, Object: models.ObjectTypeItem})
	if err != nil {
		return fmt.Errorf("error getting object after attachment deletion: %w", err)
	}

	for k, v := range resObj.Attachments {
		if v.ID == attachmentId {
			resObj.Attachments = append(resObj.Attachments[:k], resObj.Attachments[k+1:]...)
			break
		}
	}

	v.storeObject(ctx, *resObj)

	if v.syncAfterWrite {
		err := v.sync(ctx)
		if err != nil {
			return fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := getObject(v.objectStore, *resObj)
		if err != nil {
			return fmt.Errorf("error getting object after attachment deletion (syncAfterWrite): %w", err)
		}

		return v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj, "/revisionDate")
	}

	return nil
}

func (v *webAPIVault) DeleteFolder(ctx context.Context, obj models.Folder) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return models.ErrVaultLocked
	}

	// TODO: Don't fail if object is already gone
	err := v.client.DeleteFolder(ctx, obj.ID)
	if err != nil {
		return fmt.Errorf("error deleting folder: %w", err)
	}

	v.deleteObjectFromStore(ctx, obj)

	if v.syncAfterWrite {
		return v.sync(ctx)
	}
	return nil
}

func (v *webAPIVault) DeleteGroup(ctx context.Context, obj models.Group) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return models.ErrVaultLocked
	}

	err := v.client.DeleteGroup(ctx, obj)
	if err != nil {
		return fmt.Errorf("error deleting group: %w", err)
	}

	v.deleteObjectFromStore(ctx, obj)

	return nil
}

func (v *webAPIVault) DeleteOrganizationCollection(ctx context.Context, obj models.OrgCollection) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return models.ErrVaultLocked
	}

	// TODO: Don't fail if object is already gone
	err := v.client.DeleteOrganizationCollection(ctx, obj.OrganizationID, obj.ID)
	if err != nil {
		return fmt.Errorf("error deleting organization collection: %w", err)
	}

	v.deleteObjectFromStore(ctx, obj)
	v.invalidateCollectionCache(ctx, obj.OrganizationID)

	if v.syncAfterWrite {
		return v.sync(ctx)
	}
	return nil
}

func (v *webAPIVault) DeleteItem(ctx context.Context, obj models.Item) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return models.ErrVaultLocked
	}

	// TODO: Don't fail if object is already gone
	err := v.client.DeleteObject(ctx, obj.ID)
	if err != nil {
		return fmt.Errorf("error deleting item: %w", err)
	}

	v.deleteObjectFromStore(ctx, obj)

	if v.syncAfterWrite {
		return v.sync(ctx)
	}
	return nil
}

func (v *webAPIVault) EditFolder(ctx context.Context, obj models.Folder) (*models.Folder, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	var resObj *models.Folder
	encObj, err := encryptFolder(ctx, obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
	if err != nil {
		return nil, fmt.Errorf("error encrypting folder for edition: %w", err)
	}

	resFolder, err := v.client.EditFolder(ctx, *encObj)
	if err != nil {
		return nil, fmt.Errorf("error editing folder: %w", err)
	}

	resObj, err = decryptFolder(*resFolder, v.loginAccount.Secrets)
	if err != nil {
		return nil, fmt.Errorf("error decrypting folder after edition: %w", err)
	}

	v.storeObject(ctx, *resObj)

	if v.syncAfterWrite {
		err := v.sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := getObject(v.objectStore, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting folder after edition (sync-after-write): %w", err)
		}

		// NOTE: The official Bitwarden server returns dates that are a few milliseconds apart
		//       between the object's creation call and a later retrieval. We need to ignore
		//       these differences in the diff.
		return remoteObj, v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj, "/revisionDate")
	}
	return resObj, nil
}

func (v *webAPIVault) EditGroup(ctx context.Context, obj models.Group) (*models.Group, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	resObj, err := v.client.EditGroup(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("error editing group: %w", err)
	}

	if resObj.Collections == nil {
		resObj.Collections = []models.OrgCollectionMember{}
	}

	if v.syncAfterWrite {
		remoteObj, err := v.getGroup(ctx, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting group after edition (sync-after-write): %w", err)
		}
		return remoteObj, v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj)
	}
	return resObj, nil
}

func (v *webAPIVault) EditItem(ctx context.Context, obj models.Item) (*models.Item, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	// Special handling for collections identifiers changes, since you need to
	// call a different endpoint to update them.
	currentObj, err := getObject(v.objectStore, obj)
	if err != nil {
		return nil, fmt.Errorf("error getting item prior to edition: %w", err)
	}

	if !slices.Equal(currentObj.CollectionIds, obj.CollectionIds) {
		_, err = v.client.EditItemCollections(ctx, obj.ID, obj.CollectionIds)
		if err != nil {
			return nil, fmt.Errorf("error editing item collections: %w", err)
		}
	}

	var resObj *models.Item
	encObj, err := encryptItem(ctx, obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item for edition: %w", err)
	}

	resObj, err = v.client.EditItem(ctx, *encObj)
	if err != nil {
		return nil, fmt.Errorf("error editing item: %w", err)
	}

	resObj, err = decryptItem(*resObj, v.loginAccount.Secrets)
	if err != nil {
		return nil, fmt.Errorf("error decrypting item after edition: %w", err)
	}

	v.storeObject(ctx, *resObj)

	if v.syncAfterWrite {
		err := v.sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := getObject(v.objectStore, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting item after edition (sync-after-write): %w", err)
		}

		// NOTE: The official Bitwarden server returns dates that are a few milliseconds apart
		//       between the object's creation call and a later retrieval. We need to ignore
		//       these differences in the diff.
		// NOTE: The official Bitwarden server don't return the collectionIds in the response
		//       for items, even if they're actually taken into account.
		return remoteObj, v.verifyObjectAfterWrite(ctx, *resObj, *remoteObj, "/revisionDate", "/collectionIds")
	}
	return resObj, nil
}

func (v *webAPIVault) GetAPIKey(ctx context.Context, username, password string) (*models.ApiKey, error) {
	resp, err := v.client.GetAPIKey(ctx, username, password, v.loginAccount.KdfConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting API key: %w", err)
	}

	apiKey := &models.ApiKey{
		ClientID:     fmt.Sprintf("user.%s", v.loginAccount.AccountUUID),
		ClientSecret: resp.ApiKey,
	}

	return apiKey, nil
}

func (v *webAPIVault) GetAttachment(ctx context.Context, itemId, attachmentId string) ([]byte, error) {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if !v.objectsLoaded() {
		return nil, models.ErrVaultLocked
	}

	res, err := v.client.GetCipherAttachment(ctx, itemId, attachmentId)
	if err != nil {
		if strings.Contains(err.Error(), "Attachment doesn't exist") {
			return nil, models.ErrAttachmentNotFound
		}
		if strings.Contains(err.Error(), "Cipher doesn't exist") {
			return nil, models.ErrObjectNotFound
		}
		return nil, fmt.Errorf("error getting attachment object: %w", err)
	}

	rawBody, err := v.client.GetContentFromURL(ctx, res.Url)
	if err != nil {
		return nil, fmt.Errorf("error fetching attachment body: %w", err)
	}
	originalObj, err := getObject(v.objectStore, models.Item{ID: itemId, Object: models.ObjectTypeItem})
	if err != nil {
		return nil, fmt.Errorf("error getting original object: %w", err)
	}

	objectKey, err := v.getOrDefaultObjectKey(*originalObj)
	if err != nil {
		return nil, fmt.Errorf("error decoding attachment object key: %w", err)
	}

	attachmentKey, err := decryptStringAsKey(res.Key, *objectKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting attachment object key: %w", err)
	}

	encBody, err := encryptedstring.NewFromEncryptedBuffer(rawBody)
	if err != nil {
		return nil, fmt.Errorf("error creating encrypted string from attachment body: %w", err)
	}

	decryptedBody, err := crypto.Decrypt(encBody, attachmentKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting attachment body: %w", err)
	}

	return []byte(decryptedBody), nil
}

func (v *webAPIVault) GetGroup(ctx context.Context, obj models.Group) (*models.Group, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return v.getGroup(ctx, obj)
}

func (v *webAPIVault) GetOrganizationMember(ctx context.Context, obj models.OrgMember) (*models.OrgMember, error) {
	// Write lock is needed since we eventually load the organization members.
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	err := v.ensureUsersLoadedForOrg(ctx, obj.OrganizationId)
	if err != nil {
		return nil, fmt.Errorf("error loading users of organization '%s': %w", obj.OrganizationId, err)
	}

	return v.organizationMembers.FindMemberByID(obj.OrganizationId, obj.ID)
}

func (v *webAPIVault) GetOrganizationCollection(ctx context.Context, obj models.OrgCollection) (*models.OrgCollection, error) {
	// Write lock is needed since we eventually load collections.
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	return v.getOrganizationCollection(ctx, obj)
}

func (v *webAPIVault) getOrganizationCollection(ctx context.Context, obj models.OrgCollection) (*models.OrgCollection, error) {
	err := v.ensureCollectionLoadedForOrg(ctx, obj.OrganizationID)
	if err != nil {
		// We do our best to load the collection details, but we don't want to fail if we can't.
		// We'll just miss membership information.
		tflog.Error(ctx, "error loading collections details", map[string]interface{}{"error": err, "id": obj.ID})
	}

	return getObject(v.objectStore, obj)
}

func (v *webAPIVault) InviteUser(ctx context.Context, orgId, userEmail string, memberRoleType models.OrgMemberRoleType) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	req := webapi.InviteUserRequest{
		Emails:               []string{userEmail},
		Type:                 memberRoleType,
		AccessAll:            false, // TODO: Make this configurable
		AccessSecretsManager: false,
		Groups:               []string{},
		Collections:          []string{},
	}

	v.organizationMembers.ForgetOrganization(orgId)

	return v.client.InviteUser(ctx, orgId, req)
}

func (v *webAPIVault) IsSyncAfterWriteVerificationDisabled() bool {
	return !v.failOnSyncAfterWriteVerification
}

func (v *webAPIVault) FindOrganizationMember(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgMember, error) {
	filter := bitwarden.ListObjectsOptionsToFilterOptions(options...)
	if !filter.HasSearchFilter() {
		return nil, fmt.Errorf("missing search filter")
	}

	// Write lock is needed since we eventually load the organization members.
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	orgId := filter.OrganizationFilter
	userEmail := filter.SearchFilter
	err := v.ensureUsersLoadedForOrg(ctx, orgId)
	if err != nil {
		return nil, fmt.Errorf("error loading users of organization '%s': %w", orgId, err)
	}

	return v.organizationMembers.FindMemberByEmail(orgId, userEmail)
}

func (v *webAPIVault) FindOrganizationCollection(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgCollection, error) {
	filter := bitwarden.ListObjectsOptionsToFilterOptions(options...)
	if !filter.HasSearchFilter() {
		return nil, fmt.Errorf("missing search filter")
	}

	// Write lock is needed since we eventually load collections.
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	err := v.ensureCollectionLoadedForOrg(ctx, filter.OrganizationFilter)
	if err != nil {
		// We do our best to load the collection details, but we don't want to fail if we can't.
		// We'll just miss membership information.
		tflog.Error(ctx, "error loading collections details", map[string]interface{}{"error": err})
	}

	return findObject[models.OrgCollection](ctx, v.objectStore, models.ObjectTypeOrgCollection, options...)
}

func (v *webAPIVault) LoginWithAPIKey(ctx context.Context, password, clientId, clientSecret string) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if v.loginAccount.LoggedIn() {
		return models.ErrAlreadyLoggedIn
	}

	tokenResp, err := v.client.LoginWithAPIKey(ctx, clientId, clientSecret)
	if err != nil {
		return fmt.Errorf("error login with api key: %w", err)
	}

	return v.continueLoginWithTokens(ctx, *tokenResp, password)
}

func (v *webAPIVault) LoginWithPassword(ctx context.Context, username, password string) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	if v.loginAccount.LoggedIn() {
		return models.ErrAlreadyLoggedIn
	}

	preResp, err := v.client.PreLogin(ctx, username)
	if err != nil {
		return fmt.Errorf("error prelogin with username/password: %w", err)
	}

	kdfConfig := models.KdfConfiguration{
		KdfType:        preResp.Kdf,
		KdfIterations:  preResp.KdfIterations,
		KdfMemory:      preResp.KdfMemory,
		KdfParallelism: preResp.KdfParallelism,
	}

	tokenResp, err := v.client.LoginWithPassword(ctx, username, password, kdfConfig)
	if err != nil {
		return fmt.Errorf("error login with username/password: %w", err)
	}

	return v.continueLoginWithTokens(ctx, *tokenResp, password)
}

func (v *webAPIVault) Logout(ctx context.Context) error {
	v.client.ClearSession()
	v.clearObjectStore(ctx)
	v.loginAccount = Account{}
	return nil
}

func (v *webAPIVault) RegisterUser(ctx context.Context, name, username, password string, kdfConfig models.KdfConfiguration) error {
	preloginKey, err := keybuilder.BuildPreloginKey(password, username, kdfConfig)
	if err != nil {
		return fmt.Errorf("error building prelogin key: %w", err)
	}

	hashedPassword := crypto.HashPassword(password, *preloginKey, false)

	encryptionKey, encryptedEncryptionKey, err := keybuilder.GenerateEncryptionKey(*preloginKey)
	if err != nil {
		return fmt.Errorf("error generating encryption key: %w", err)
	}

	publicKey, encryptedPrivateKey, err := keybuilder.GenerateEncryptedRSAKeyPair(*encryptionKey)
	if err != nil {
		return fmt.Errorf("error generating key pair: %w", err)
	}

	signupRequest := webapi.SignupRequest{
		Email:              username,
		Name:               name,
		MasterPasswordHash: hashedPassword,
		Key:                encryptedEncryptionKey,
		Kdf:                kdfConfig.KdfType,
		KdfIterations:      kdfConfig.KdfIterations,
		KdfMemory:          kdfConfig.KdfMemory,
		KdfParallelism:     kdfConfig.KdfParallelism,
		Keys: webapi.KeyPair{
			PublicKey:           publicKey,
			EncryptedPrivateKey: encryptedPrivateKey,
		},
	}

	return v.client.RegisterUser(ctx, signupRequest)
}

func (v *webAPIVault) Sync(ctx context.Context) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	return v.sync(ctx)
}

func (v *webAPIVault) sync(ctx context.Context) error {
	if !v.loginAccount.LoggedIn() {
		return models.ErrLoggedOut
	} else if !v.loginAccount.SecretsLoaded() {
		return models.ErrVaultLocked
	}

	ciphersRaw, err := v.client.Sync(ctx)
	if err != nil {
		return fmt.Errorf("error syncing: %w", err)
	}

	if v.loginAccount.Email != ciphersRaw.Profile.Email || v.loginAccount.AccountUUID != ciphersRaw.Profile.Id {
		return fmt.Errorf("BUG: account UUID or email changed during sync")
	}

	err = loadOrganizationSecrets(v.loginAccount.Secrets, ciphersRaw.Profile.Organizations)
	if err != nil {
		return fmt.Errorf("error loading organization secrets: %w", err)
	}

	return v.loadObjectMap(ctx, *ciphersRaw)
}

func (v *webAPIVault) Unlock(ctx context.Context, password string) error {
	v.vaultOperationMutex.Lock()
	defer v.vaultOperationMutex.Unlock()

	return v.unlock(ctx, password)
}

func (v *webAPIVault) unlock(ctx context.Context, password string) error {
	if !v.loginAccount.LoggedIn() {
		return models.ErrLoggedOut
	}

	profile, err := v.client.GetProfile(ctx)
	if err != nil {
		return fmt.Errorf("error loading profile: %w", err)
	}

	v.loginAccount.Email = profile.Email
	v.loginAccount.AccountUUID = profile.Id

	accountSecrets, err := decryptAccountSecrets(v.loginAccount, password)
	if err != nil {
		return fmt.Errorf("error decrypting account secrets: %w", err)
	}
	v.loginAccount.Secrets = *accountSecrets

	return nil
}

func (v *webAPIVault) continueLoginWithTokens(ctx context.Context, tokenResp webapi.TokenResponse, password string) error {
	v.loginAccount = Account{
		VaultFormat: "API",
		KdfConfig: models.KdfConfiguration{
			KdfType:        tokenResp.Kdf,
			KdfIterations:  tokenResp.KdfIterations,
			KdfMemory:      tokenResp.KdfMemory,
			KdfParallelism: tokenResp.KdfParallelism,
		},
		ProtectedRSAPrivateKey: tokenResp.PrivateKey,
		ProtectedSymmetricKey:  tokenResp.Key,
	}

	err := v.unlock(ctx, password)
	if err != nil {
		return fmt.Errorf("error unlocking after login: %w", err)
	}

	return v.sync(ctx)
}

func (v *webAPIVault) getGroup(ctx context.Context, obj models.Group) (*models.Group, error) {
	return v.client.GetGroup(ctx, obj)
}

func (v *webAPIVault) getUserPublicKey(ctx context.Context, userId string) (*rsa.PublicKey, error) {
	userPublicKey, err := v.client.GetUserPublicKey(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting user's public key: %w", err)
	}

	decodedKey, err := base64.StdEncoding.DecodeString(string(userPublicKey))
	if err != nil {
		return nil, fmt.Errorf("error decoding public key: %w", err)
	}

	pubKey, err := x509.ParsePKIXPublicKey(decodedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}
	return pubKey.(*rsa.PublicKey), nil
}

func (v *webAPIVault) reloadCollectionDetailsOfOrg(ctx context.Context, orgId string) error {
	v.invalidateCollectionCache(ctx, orgId)
	return v.ensureCollectionLoadedForOrg(ctx, orgId)
}

func (v *webAPIVault) ensureCollectionLoadedForOrg(ctx context.Context, orgId string) error {
	if _, ok := v.collectionDetailsLoadedForOrg[orgId]; ok {
		return nil
	}

	tflog.Trace(ctx, "Loading collections for organization", map[string]interface{}{"org_id": orgId})
	accessDetails, err := v.client.GetOrganizationCollections(ctx, orgId)
	if err != nil {
		return fmt.Errorf("error reading collection details: %w", err)
	}

	for _, collection := range accessDetails {
		orgCol, err := decryptOrgCollection(collection, v.loginAccount.Secrets)
		if err != nil {
			return fmt.Errorf("error decrypting collection: %w", err)
		}

		v.storeObject(ctx, *orgCol)
	}
	return nil
}

func (v *webAPIVault) invalidateCollectionCache(ctx context.Context, orgId string) {
	tflog.Trace(ctx, "Invalidating cached collections for organization", map[string]interface{}{"org_id": orgId})
	delete(v.collectionDetailsLoadedForOrg, orgId)
}

func (v *webAPIVault) prepareAttachmentCreationRequest(_ context.Context, itemId, filename string, content []byte) (*webapi.AttachmentRequestData, []byte, error) {
	// NOTE: We don't Sync() to get the latest version of Object before adding an attachment to it, because we
	//       assume the Object's key can't change.
	originalObj, err := getObject(v.objectStore, models.Item{ID: itemId, Object: models.ObjectTypeItem})
	if err != nil {
		return nil, nil, fmt.Errorf("error getting original object: %w", err)
	}

	objectKey, err := v.getOrDefaultObjectKey(*originalObj)
	if err != nil {
		return nil, nil, fmt.Errorf("error get cipher key while creating attachment: %w", err)
	}

	attachmentKey, err := keybuilder.CreateObjectKey()
	if err != nil {
		return nil, nil, err
	}

	encData, err := crypto.Encrypt(content, *attachmentKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error encrypting data: %w", err)
	}

	encDataBuffer, err := encData.ToEncryptedBuffer()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting encrypted buffer: %w", err)
	}

	encFilename, err := crypto.EncryptAsString([]byte(filename), *objectKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error encrypting filename: %w", err)
	}

	dataKeyEncrypted, err := crypto.EncryptAsString(attachmentKey.Key, *objectKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error encrypting dataKeyEncrypted: %w", err)
	}

	req := webapi.AttachmentRequestData{
		FileName: encFilename,
		FileSize: len(encDataBuffer),
		Key:      dataKeyEncrypted,
	}
	return &req, encDataBuffer, nil
}

func (v *webAPIVault) loadObjectMap(ctx context.Context, cipherMap webapi.SyncResponse) error {
	v.clearObjectStore(ctx)

	v.storeOrganizationSecrets(ctx)

	res, err := ciphersToObjects(v.loginAccount.Secrets, cipherMap.Ciphers)
	if err != nil {
		return fmt.Errorf("error updating object in store: %w", err)
	}
	v.storeObjects(ctx, res)

	res, err = ciphersToObjects(v.loginAccount.Secrets, cipherMap.Folders)
	if err != nil {
		return fmt.Errorf("error updating folder in store: %w", err)
	}
	v.storeObjects(ctx, res)

	res, err = ciphersToObjects(v.loginAccount.Secrets, cipherMap.Collections)
	if err != nil {
		return fmt.Errorf("error updating org collection in store: %w", err)
	}
	v.storeObjects(ctx, res)

	return nil
}

func (v *webAPIVault) verifyObjectAfterWrite(ctx context.Context, actual, expected interface{}, ignoreFields ...string) error {
	err := compareObjects(ctx, actual, expected, ignoreFields...)
	if err != nil {
		tflog.Error(ctx, "server returned different object after write", map[string]interface{}{"error": err})

		if v.failOnSyncAfterWriteVerification {
			return fmt.Errorf(`server returned different object after write!
After writing an object and re-fetching it, the server returned a slightly different version: %w

To learn more about this issue and how to handle it, please:
1. Consider reporting affected fields at: https://github.com/maxlaverse/terraform-provider-bitwarden/issues/new
2. Check the documentation of the 'experimental.disable_sync_after_write_verification' attribute`, err)
		}
	}
	return nil
}

func (v *webAPIVault) checkMembersExistence(ctx context.Context, orgId string, users []models.OrgCollectionMember) error {
	err := v.ensureUsersLoadedForOrg(ctx, orgId)
	if err != nil {
		return fmt.Errorf("error getting users for org: %w", err)
	}

	for _, user := range users {
		_, err := v.organizationMembers.FindMemberByID(orgId, user.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

type SupportedCipher interface {
	models.Item | models.Folder | webapi.Collection
}

func ciphersToObjects[T SupportedCipher](accountSecrets AccountSecrets, ciphers []T) ([]interface{}, error) {
	objects := make([]interface{}, len(ciphers))
	for k, value := range ciphers {
		switch secret := any(value).(type) {
		case models.Item:
			obj, err := decryptItem(secret, accountSecrets)
			if err != nil {
				return nil, fmt.Errorf("error decrypting cipher item: %w", err)
			}
			objects[k] = *obj
		case models.Folder:
			obj, err := decryptFolder(secret, accountSecrets)
			if err != nil {
				return nil, fmt.Errorf("error decrypting cipher folder: %w", err)
			}
			objects[k] = *obj
		case webapi.Collection:
			obj, err := decryptOrgCollection(secret, accountSecrets)
			if err != nil {
				return nil, fmt.Errorf("error decrypting cipher collection: %w", err)
			}
			objects[k] = *obj
		default:
			return nil, fmt.Errorf("BUG: ciphersToObjects, unknown object type: %T", value)
		}
	}
	return objects, nil
}

func loadOrganizationSecrets(accountSecrets AccountSecrets, organizations []webapi.Organization) error {
	for _, organization := range organizations {
		key, err := decryptOrganizationKey(organization.Key, accountSecrets.RSAPrivateKey)
		if err != nil {
			return fmt.Errorf("error loading organization key: %w", err)
		}

		orgSecret := OrganizationSecret{
			OrganizationUUID: organization.Id,
			Key:              *key,
			Name:             organization.Name,
		}
		accountSecrets.OrganizationSecrets[orgSecret.OrganizationUUID] = orgSecret

	}
	return nil
}

func checkForDuplicateMembers(users []models.OrgCollectionMember) error {
	uniqueMembers := make(map[string]int)
	for _, member := range users {
		uniqueMembers[member.Id]++
	}

	for memberId, count := range uniqueMembers {
		if count > 1 {
			return fmt.Errorf("member ID '%s' was specified twice", memberId)
		}
	}
	return nil
}
