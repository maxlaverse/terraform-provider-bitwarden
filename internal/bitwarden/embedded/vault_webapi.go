package embedded

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

type WebAPIVault interface {
	BaseVault
	CreateObject(ctx context.Context, obj models.Object) (*models.Object, error)
	CreateOrganization(ctx context.Context, organizationName, organizationLabel, billingEmail string) (string, error)
	CreateAttachment(ctx context.Context, itemId, filePath string) (*models.Object, error)
	DeleteAttachment(ctx context.Context, itemId, attachmentId string) error
	DeleteObject(ctx context.Context, obj models.Object) error
	EditObject(ctx context.Context, obj models.Object) (*models.Object, error)
	GetAPIKey(ctx context.Context, username, password string) (*models.ApiKey, error)
	GetAttachment(ctx context.Context, itemId, attachmentId string) ([]byte, error)
	LoginWithAPIKey(ctx context.Context, password, clientId, clientSecret string) error
	LoginWithPassword(ctx context.Context, username, password string) error
	RegisterUser(ctx context.Context, name, username, password string, kdfConfig models.KdfConfiguration) error
	Sync(ctx context.Context) error
	Unlock(ctx context.Context, password string) error
}

type Options func(c bitwarden.Client)

// DisableCryptoSafeMode disables the safe mode for crypto operations, which reverses
// crypto.Encrypt() to make sure it can decrypt the result.
func DisableCryptoSafeMode() Options {
	return func(c bitwarden.Client) {
		crypto.SafeMode = false
	}
}

// DisableObjectEncryptionVerification disables the systematic attempts to decrypt objects
// (items, folders, collections) after they have been created or edited, to verify that the
// encryption can be reverse.
func DisableObjectEncryptionVerification() Options {
	return func(c bitwarden.Client) {
		c.(*webAPIVault).baseVault.verifyObjectEncryption = false
	}
}

// DisableSyncAfterWrite disables the systematic Sync() after a write operation (create, edit,
// delete) to the vault. Write operations already return the object that was created or edited, so
// Sync() is not strictly necessary.
func DisableSyncAfterWrite() Options {
	return func(c bitwarden.Client) {
		c.(*webAPIVault).syncAfterWrite = false
	}
}

// DisableRetryBackoff disables the retry backoff mechanism for API calls.
func WithHttpOptions(opts ...webapi.Options) Options {
	return func(c bitwarden.Client) {
		c.(*webAPIVault).client = webapi.NewClient(c.(*webAPIVault).serverURL, opts...)
	}
}

// Panic on error is useful for debugging, but should not be used in production.
func EnablePanicOnEncryptionError() Options {
	return func(c bitwarden.Client) {
		panicOnEncryptionErrors = true
	}
}

func NewWebAPIVault(serverURL string, opts ...Options) WebAPIVault {
	c := &webAPIVault{
		baseVault: baseVault{
			objectStore:            make(map[string]models.Object),
			locked:                 true,
			verifyObjectEncryption: true,
		},
		serverURL: serverURL,

		// Always run Sync() after creating, editing, or deleting an object and verify the result
		// by comparing the local and remote objects.
		syncAfterWrite: true,
	}

	for _, o := range opts {
		o(c)
	}

	if c.client == nil {
		c.client = webapi.NewClient(serverURL)
	}

	return c
}

func NewDeviceIdentifier() string {
	return uuid.New().String()
}

type webAPIVault struct {
	baseVault
	client webapi.Client

	ciphersMap     webapi.SyncResponse
	syncAfterWrite bool
	serverURL      string
}

func (v *webAPIVault) CreateAttachment(ctx context.Context, itemId, filePath string) (*models.Object, error) {
	if v.locked {
		return nil, models.ErrVaultLocked
	}

	req, data, err := v.prepareAttachmentCreationRequest(ctx, itemId, filePath)
	if err != nil {
		return nil, fmt.Errorf("error preparing attachment creation request: %w", err)
	}

	resp, err := v.client.CreateObjectAttachment(ctx, itemId, data, *req)
	if err != nil {
		return nil, fmt.Errorf("error creating attachment: %w", err)
	}

	err = v.client.CreateObjectAttachmentData(ctx, itemId, resp.AttachmentId, data)
	if err != nil {
		return nil, fmt.Errorf("error creating attachment data: %w", err)
	}

	resObj, err := decryptItem((*resp).CipherResponse, v.loginAccount.Secrets)
	if err != nil {
		return nil, fmt.Errorf("error decrypting resulting obj data attachment: %w", err)
	}

	v.storeObject(ctx, *resObj)

	if v.syncAfterWrite {
		err = v.Sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := v.GetObject(ctx, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting object after attachment upload (sync-after-write): %w", err)
		}

		// The attachment's URL contains a signed token generated on each request. We need to diff
		// it out if we want the comparison to work.
		for k, v := range remoteObj.Attachments {
			resObj.Attachments[k].Url = v.Url
		}

		return remoteObj, compareObjects(*resObj, *remoteObj)
	}
	return resObj, nil
}

func (v *webAPIVault) CreateObject(ctx context.Context, obj models.Object) (*models.Object, error) {
	if v.locked {
		return nil, models.ErrVaultLocked
	}

	var resObj *models.Object
	if obj.Object == models.ObjectTypeFolder {
		encObj, err := v.encryptFolder(ctx, obj, v.loginAccount.Secrets)
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
	} else if obj.Object == models.ObjectTypeOrgCollection {
		encObj, err := encryptCollection(obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
		if err != nil {
			return nil, fmt.Errorf("error encrypting collection for creation: %w", err)
		}

		resEncCollection, err := v.client.CreateOrgCollection(ctx, obj.OrganizationID, *encObj)
		if err != nil {
			return nil, fmt.Errorf("error creating collection: %w", err)
		}

		resObj, err = decryptCollection(*resEncCollection, v.loginAccount.Secrets)
		if err != nil {
			return nil, fmt.Errorf("error decrypting collection after creation: %w", err)
		}
	} else {
		encObj, err := encryptItem(ctx, obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
		if err != nil {
			return nil, fmt.Errorf("error encrypting item for creation: %w", err)
		}

		resEncObj, err := v.client.CreateObject(ctx, *encObj)

		if err != nil {
			return nil, fmt.Errorf("error creating item: %w", err)
		}

		resEncObj.Object = obj.Object
		resEncObj.Type = obj.Type
		resObj, err = decryptItem(*resEncObj, v.loginAccount.Secrets)
		if err != nil {
			return nil, fmt.Errorf("error decrypting item after creation: %w", err)
		}
	}

	v.storeObject(ctx, *resObj)
	if v.syncAfterWrite {
		err := v.Sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := v.GetObject(ctx, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting object after creation (sync-after-write): %w", err)
		}

		// NOTE: The official Bitwarden server returns dates that are a few milliseconds apart
		//       between the object's creation call and a later retrieval. We need to ignore
		//       these differences in the diff.
		resObj.CreationDate = remoteObj.CreationDate
		resObj.RevisionDate = remoteObj.RevisionDate

		return remoteObj, compareObjects(*resObj, *remoteObj)
	}
	return resObj, nil
}

func (v *webAPIVault) CreateOrganization(ctx context.Context, organizationName, organizationLabel, billingEmail string) (string, error) {
	if v.locked {
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

	publicKey, encryptedPrivateKey, err := keybuilder.GenerateRSAKeyPair(*sharedKey)
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
	return res.Id, nil
}

func (v *webAPIVault) DeleteAttachment(ctx context.Context, itemId, attachmentId string) error {
	// TODO: Don't fail if attachment is already gone
	err := v.client.DeleteObjectAttachment(ctx, itemId, attachmentId)
	if err != nil {
		return fmt.Errorf("error deleting attachment: %w", err)
	}

	resObj, err := v.GetObject(ctx, models.Object{ID: itemId, Object: models.ObjectTypeItem})
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
		err := v.Sync(ctx)
		if err != nil {
			return fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := v.GetObject(ctx, *resObj)
		if err != nil {
			return fmt.Errorf("error getting object after attachment deletion (syncAfterWrite): %w", err)
		}

		return compareObjects(*resObj, *remoteObj)
	}

	return nil
}

func (v *webAPIVault) DeleteObject(ctx context.Context, obj models.Object) error {
	// TODO: Don't fail if object is already gone
	var err error
	if obj.Object == models.ObjectTypeFolder {
		err = v.client.DeleteFolder(ctx, obj.ID)
	} else if obj.Object == models.ObjectTypeOrgCollection {
		err = v.client.DeleteOrgCollection(ctx, obj.OrganizationID, obj.ID)
	} else {
		err = v.client.DeleteObject(ctx, obj.ID)
	}

	if err != nil {
		return fmt.Errorf("error deleting object: %w", err)
	}

	v.deleteObjectFromStore(ctx, obj)

	if v.syncAfterWrite {
		return v.Sync(ctx)
	}
	return nil
}

func (v *webAPIVault) EditObject(ctx context.Context, obj models.Object) (*models.Object, error) {
	if v.locked {
		return nil, models.ErrVaultLocked
	}

	var resObj *models.Object
	if obj.Object == models.ObjectTypeFolder {
		encObj, err := v.encryptFolder(ctx, obj, v.loginAccount.Secrets)
		if err != nil {
			return nil, fmt.Errorf("error encrypting folder for edition: %w", err)
		}

		resFolder, err := v.client.EditFolder(ctx, *encObj)
		if err != nil {
			return nil, fmt.Errorf("error editing folder: %w", err)
		}

		resObj, err = decryptFolder(*resFolder, v.loginAccount.Secrets)
		if err != nil {
			return nil, fmt.Errorf("error decrypting folder after creation: %w", err)
		}
	} else if obj.Object == models.ObjectTypeOrgCollection {
		encObj, err := encryptCollection(obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
		if err != nil {
			return nil, fmt.Errorf("error encrypting collection for creation: %w", err)
		}

		resCollection, err := v.client.EditOrgCollection(ctx, obj.OrganizationID, obj.ID, *encObj)
		if err != nil {
			return nil, fmt.Errorf("error editing collection: %w", err)
		}

		resObj, err = decryptCollection(*resCollection, v.loginAccount.Secrets)
		if err != nil {
			return nil, fmt.Errorf("error decrypting collection after creation: %w", err)
		}
	} else {
		encObj, err := encryptItem(ctx, obj, v.loginAccount.Secrets, v.verifyObjectEncryption)
		if err != nil {
			return nil, fmt.Errorf("error encrypting item for edition: %w", err)
		}

		resObj, err = v.client.EditObject(ctx, *encObj)
		if err != nil {
			return nil, fmt.Errorf("error editing item: %w", err)
		}

		resObj, err = decryptItem(*resObj, v.loginAccount.Secrets)
		if err != nil {
			return nil, fmt.Errorf("error decrypting item after creation: %w", err)
		}
	}

	v.storeObject(ctx, *resObj)

	if v.syncAfterWrite {
		err := v.Sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("sync-after-write error: %w", err)
		}

		remoteObj, err := v.GetObject(ctx, *resObj)
		if err != nil {
			return nil, fmt.Errorf("error getting object after edition (sync-after-write): %w", err)
		}

		// NOTE: The official Bitwarden server returns dates that are a few milliseconds apart
		//       between the object's creation call and a later retrieval. We need to ignore
		//       these differences in the diff.
		resObj.RevisionDate = remoteObj.RevisionDate

		return remoteObj, compareObjects(*resObj, *remoteObj)
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
	if v.locked {
		return nil, models.ErrVaultLocked
	}

	res, err := v.client.GetObjectAttachment(ctx, itemId, attachmentId)
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

	originalObj, err := v.GetObject(ctx, models.Object{ID: itemId, Object: models.ObjectTypeItem})
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

func (v *webAPIVault) LoginWithAPIKey(ctx context.Context, password, clientId, clientSecret string) error {
	tokenResp, err := v.client.LoginWithAPIKey(ctx, clientId, clientSecret)
	if err != nil {
		return fmt.Errorf("error login with api key: %w", err)
	}

	return v.continueLoginWithTokens(ctx, *tokenResp, password)
}

func (v *webAPIVault) LoginWithPassword(ctx context.Context, username, password string) error {
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

	publicKey, encryptedPrivateKey, err := keybuilder.GenerateRSAKeyPair(*encryptionKey)
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
	ciphersRaw, err := v.client.Sync(ctx)
	if err != nil {
		return fmt.Errorf("error syncing: %w", err)
	}
	if len(v.loginAccount.Email) > 0 && v.loginAccount.Email != ciphersRaw.Profile.Email || len(v.loginAccount.AccountUUID) > 0 && v.loginAccount.AccountUUID != ciphersRaw.Profile.Id {
		return fmt.Errorf("BUG: account UUID or email changed during sync")
	}

	v.ciphersMap = *ciphersRaw
	v.loginAccount.Email = v.ciphersMap.Profile.Email
	v.loginAccount.AccountUUID = v.ciphersMap.Profile.Id

	if !v.loginAccount.PrivateKeyDecrypted() {
		return nil
	}

	return v.loadObjectMap(ctx)
}

func (v *webAPIVault) Unlock(ctx context.Context, password string) error {
	if len(v.loginAccount.Email) == 0 {
		return fmt.Errorf("please login first")
	}

	accountSecrets, err := decryptAccountSecrets(v.loginAccount, password)
	if err != nil {
		return fmt.Errorf("error decrypting account secrets: %w", err)
	}
	v.loginAccount.Secrets = *accountSecrets

	profile, err := v.client.Profile(ctx)
	if err != nil {
		return fmt.Errorf("error loading profile: %w", err)
	}

	v.ciphersMap.Profile = *profile

	err = v.loadObjectMap(ctx)
	if err != nil {
		return fmt.Errorf("error loading cipher map: %w", err)
	}

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

	err := v.Sync(ctx)
	if err != nil {
		return fmt.Errorf("error syncing after login: %w", err)
	}

	return v.Unlock(ctx, password)
}

func (v *webAPIVault) loadCollectionsFromObjectMap(ctx context.Context) error {
	for _, collection := range v.ciphersMap.Collections {
		obj, err := decryptCollection(collection, v.loginAccount.Secrets)
		if err != nil {
			return fmt.Errorf("error decrypting collection: %w", err)
		}
		v.storeObject(ctx, *obj)
	}
	return nil
}

func (v *webAPIVault) loadFoldersFromObjectMap(ctx context.Context) error {
	for _, folder := range v.ciphersMap.Folders {
		obj, err := decryptFolder(folder, v.loginAccount.Secrets)
		if err != nil {
			return fmt.Errorf("error decrypting folder: %w", err)
		}
		v.storeObject(ctx, *obj)
	}
	return nil
}

func (v *webAPIVault) loadObjectsFromObjectMap(ctx context.Context) error {
	for _, value := range v.ciphersMap.Ciphers {
		obj, err := decryptItem(value, v.loginAccount.Secrets)
		if err != nil {
			return fmt.Errorf("error decrypting object: %w", err)
		}
		v.storeObject(ctx, *obj)
	}
	return nil
}

func (v *webAPIVault) loadObjectMap(ctx context.Context) error {
	v.clearObjectStore(ctx)

	err := v.loadOrganizationSecretsFromObjectMap(ctx)
	if err != nil {
		return fmt.Errorf("error updating organization secrets: %w", err)
	}
	err = v.loadObjectsFromObjectMap(ctx)
	if err != nil {
		return fmt.Errorf("error updating object in store: %w", err)
	}

	err = v.loadFoldersFromObjectMap(ctx)
	if err != nil {
		return fmt.Errorf("error updating folder in store: %w", err)
	}

	err = v.loadCollectionsFromObjectMap(ctx)
	if err != nil {
		return fmt.Errorf("error updating collections in store: %w", err)
	}

	tflog.Debug(ctx, "Vault is unlocked")
	v.locked = false
	return nil
}

func (v *webAPIVault) loadOrganizationSecretsFromObjectMap(ctx context.Context) error {
	for _, organization := range v.ciphersMap.Profile.Organizations {
		key, err := decryptOrganizationKey(organization.Key, v.loginAccount.Secrets.RSAPrivateKey)
		if err != nil {
			return fmt.Errorf("error loading organization key: %w", err)
		}

		orgSecret := OrganizationSecret{
			OrganizationUUID: organization.Id,
			Key:              *key,
		}
		v.loginAccount.Secrets.OrganizationSecrets[orgSecret.OrganizationUUID] = orgSecret

		obj := models.Object{
			ID:     organization.Id,
			Object: models.ObjectTypeOrganization,
			Name:   organization.Name,
		}
		v.storeObject(ctx, obj)
	}
	return nil
}

func (v *webAPIVault) prepareAttachmentCreationRequest(ctx context.Context, itemId, filePath string) (*webapi.AttachmentRequestData, []byte, error) {
	// NOTE: We don't Sync() to get the latest version of Object before adding an attachment to it, because we
	//       assume the Object's key can't change.
	originalObj, err := v.GetObject(ctx, models.Object{ID: itemId, Object: models.ObjectTypeItem})
	if err != nil {
		return nil, nil, fmt.Errorf("error getting original object: %w", err)
	}

	objectKey, err := v.getOrDefaultObjectKey(*originalObj)
	if err != nil {
		return nil, nil, fmt.Errorf("error get cipher key while creating attachment: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading file: %w", err)
	}

	attachmentKey, err := keybuilder.CreateObjectKey()
	if err != nil {
		return nil, nil, err
	}

	encData, err := crypto.Encrypt(data, *attachmentKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error encrypting data: %w", err)
	}

	encDataBuffer, err := encData.ToEncryptedBuffer()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting encrypted buffer: %w", err)
	}

	filename, err := crypto.EncryptAsString([]byte(filepath.Base(filePath)), *objectKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error encrypting filename: %w", err)
	}

	dataKeyEncrypted, err := crypto.EncryptAsString(attachmentKey.Key, *objectKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error encrypting dataKeyEncrypted: %w", err)
	}

	req := webapi.AttachmentRequestData{
		FileName: filename,
		FileSize: len(encDataBuffer),
		Key:      dataKeyEncrypted,
	}
	return &req, encDataBuffer, nil
}
