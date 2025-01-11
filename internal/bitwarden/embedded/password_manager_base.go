package embedded

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"golang.org/x/net/publicsuffix"
)

var (
	panicOnEncryptionErrors = false
)

type BaseVault interface {
	GetFolder(ctx context.Context, obj models.Folder) (*models.Folder, error)
	GetItem(ctx context.Context, obj models.Item) (*models.Item, error)
	GetOrganization(context.Context, models.Organization) (*models.Organization, error)
	GetOrganizationCollection(ctx context.Context, collection models.OrgCollection) (*models.OrgCollection, error)

	FindFolder(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Folder, error)
	FindItem(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Item, error)
	FindOrganization(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Organization, error)
	FindOrganizationCollection(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgCollection, error)
}

type baseVault struct {
	loginAccount Account
	objectStore  map[string]interface{}

	// vaultOperationMutex protects the objectStore and loginAccount fields
	// from concurrent access. Read operations are allowed to run concurrently,
	// but write operations are serialized. In theory we could protect the two
	// fields individually, but it's just much more easier to have a single
	// mutex for both.
	vaultOperationMutex sync.RWMutex

	// verifyObjectEncryption is a flag that can be set to true to verify that
	// every object that is encrypted can be decrypted back to its original.
	verifyObjectEncryption bool
}

func (v *baseVault) GetItem(ctx context.Context, obj models.Item) (*models.Item, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return getObject(v.objectStore, obj)
}

func (v *baseVault) GetFolder(ctx context.Context, obj models.Folder) (*models.Folder, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return getObject(v.objectStore, obj)
}

func (v *baseVault) GetOrganization(ctx context.Context, obj models.Organization) (*models.Organization, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return getObject(v.objectStore, obj)
}

func (v *baseVault) GetOrganizationCollection(ctx context.Context, obj models.OrgCollection) (*models.OrgCollection, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return getObject(v.objectStore, obj)
}

func getObject[T any](store map[string]interface{}, obj T) (*T, error) {
	if store == nil {
		return nil, models.ErrVaultLocked
	}

	var itemType models.ItemType
	switch itemObj := any(obj).(type) {
	case models.Item:
		itemType = itemObj.Type
	}

	storedObj, ok := store[objKey(obj)]
	if !ok {
		return nil, models.ErrObjectNotFound
	}

	switch itemObj := any(storedObj).(type) {
	case models.Item:
		if itemObj.DeletedDate != nil {
			return nil, models.ErrObjectNotFound
		}
		if itemType > 0 && itemObj.Type != itemType {
			return nil, models.ErrItemTypeMismatch
		}
	}

	v := storedObj.(T)
	return &v, nil
}

func (v *baseVault) FindFolder(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Folder, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return findObject[models.Folder](ctx, v.objectStore, models.ObjectTypeFolder, options...)
}

func (v *baseVault) FindItem(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Item, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return findObject[models.Item](ctx, v.objectStore, models.ObjectTypeItem, options...)
}

func (v *baseVault) FindOrganization(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.Organization, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return findObject[models.Organization](ctx, v.objectStore, models.ObjectTypeOrganization, options...)
}

func (v *baseVault) FindOrganizationCollection(ctx context.Context, options ...bitwarden.ListObjectsOption) (*models.OrgCollection, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return findObject[models.OrgCollection](ctx, v.objectStore, models.ObjectTypeOrgCollection, options...)
}

func findObject[T any](ctx context.Context, store map[string]interface{}, objType models.ObjectType, options ...bitwarden.ListObjectsOption) (*T, error) {
	if store == nil {
		return nil, models.ErrVaultLocked
	}

	filter := bitwarden.ListObjectsOptionsToFilterOptions(options...)
	if !filter.IsValid() {
		return nil, fmt.Errorf("invalid filter options")
	}

	foundObjects := []T{}
	for _, rawObj := range store {
		obj, ok := rawObj.(T)
		if !ok {
			continue
		}

		if !objMatchFilter(ctx, obj, filter, objType) {
			continue
		}

		foundObjects = append(foundObjects, obj)
	}

	if len(foundObjects) == 0 {
		return nil, models.ErrNoObjectFoundMatchingFilter
	} else if len(foundObjects) > 1 {
		return nil, models.ErrTooManyObjectsFound
	}

	return &foundObjects[0], nil
}

func (v *baseVault) clearObjectStore(ctx context.Context) {
	if v.objectStore != nil {
		tflog.Trace(ctx, "Clearing object store")
	}
	v.objectStore = make(map[string]interface{})
}

func (v *baseVault) deleteObjectFromStore(ctx context.Context, obj any) {
	tflog.Trace(ctx, "Deleting object from store", map[string]interface{}{"key": objKey(obj)})
	delete(v.objectStore, objKey(obj))
}

func (v *baseVault) objectsLoaded() bool {
	return v.objectStore != nil
}

func (v *baseVault) storeObject(ctx context.Context, obj any) {
	tflog.Trace(ctx, "Storing new object", map[string]interface{}{"key": objKey(obj)})
	v.objectStore[objKey(obj)] = obj
}

func (v *baseVault) storeObjects(ctx context.Context, obj []interface{}) {
	for _, o := range obj {
		v.storeObject(ctx, o)
	}
}

func (v *baseVault) storeOrganizationSecrets(ctx context.Context) {
	for _, orgSecret := range v.loginAccount.Secrets.OrganizationSecrets {
		v.storeObject(ctx, models.Organization{
			ID:     orgSecret.OrganizationUUID,
			Object: models.ObjectTypeOrganization,
			Name:   orgSecret.Name,
		})
	}
}

func decryptAccountSecrets(account Account, password string) (*AccountSecrets, error) {
	if len(account.Email) == 0 {
		// A common mistake is trying to decrypt account secrets without an
		// email, the content of an Account comes from two different API calls.
		return nil, fmt.Errorf("BUG: email required to decrypt account secrets")
	}

	masterKey, err := keybuilder.BuildPreloginKey(password, account.Email, account.KdfConfig)
	if err != nil {
		return nil, fmt.Errorf("error building prelogin key: %w", err)
	}

	generatedKeys, err := decryptStringAsKey(account.ProtectedSymmetricKey, masterKey.StretchKey())
	if err != nil {
		return nil, models.ErrWrongMasterPassword
	}

	rsaPrivateKey, err := crypto.DecryptPrivateKey(account.ProtectedRSAPrivateKey, *generatedKeys)
	if err != nil {
		return nil, fmt.Errorf("error decrypting private RSA key: %w", err)
	}
	return &AccountSecrets{
		OrganizationSecrets: map[string]OrganizationSecret{},
		MasterPasswordHash:  crypto.HashPassword(password, *masterKey, false),
		MainKey:             *generatedKeys,
		RSAPrivateKey:       rsaPrivateKey,
	}, nil
}

func decryptOrgCollection(obj webapi.Collection, secret AccountSecrets) (*models.OrgCollection, error) {
	orgKey, err := secret.GetOrganizationKey(obj.OrganizationId)
	if err != nil {
		return nil, err
	}

	decName, err := decryptStringIfNotEmpty(obj.Name, *orgKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting collection name: %w", err)
	}

	return &models.OrgCollection{
		ID:             obj.Id,
		Name:           decName,
		Object:         models.ObjectTypeOrgCollection,
		OrganizationID: obj.OrganizationId,
	}, nil
}

func decryptFolder(obj webapi.Folder, secret AccountSecrets) (*models.Folder, error) {
	objName, err := decryptStringIfNotEmpty(obj.Name, secret.MainKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting folder name: %w", err)
	}

	return &models.Folder{
		ID:           obj.Id,
		Name:         objName,
		Object:       models.ObjectTypeFolder,
		RevisionDate: cloneDate(obj.RevisionDate),
	}, nil
}

func decryptItem(obj models.Item, secret AccountSecrets) (*models.Item, error) {
	objectKey, err := getObjectKey(obj, secret)
	if err != nil {
		return nil, fmt.Errorf("error getting item object key: %w", err)
	}

	decFields, err := decryptItemFields(obj.Fields, *objectKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting item fields: %w", err)
	}

	decLogin, err := decryptItemLogin(obj.Login, *objectKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting item login: %w", err)
	}

	decName, err := decryptStringIfNotEmpty(obj.Name, *objectKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting item name: %v", err)
	}

	decNotes, err := decryptStringIfNotEmpty(obj.Notes, *objectKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting item notes: %v", err)
	}

	decAttachments := make([]models.Attachment, len(obj.Attachments))
	for k, f := range obj.Attachments {
		decFilename, err := decryptStringIfNotEmpty(f.FileName, *objectKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting attachment filename: %v", err)
		}

		decAttachments[k] = models.Attachment{
			FileName: decFilename,
			ID:       f.ID,
			Key:      f.Key,
			Object:   f.Object,
			Size:     f.Size,
			SizeName: f.SizeName,
			Url:      f.Url,
		}
	}

	decPasswordHistory := make([]models.PasswordHistoryItem, len(obj.PasswordHistory))
	for k, f := range obj.PasswordHistory {
		decPassword, err := decryptStringIfNotEmpty(f.Password, *objectKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting attachment filename: %v", err)
		}

		decPasswordHistory[k] = models.PasswordHistoryItem{
			Password:     decPassword,
			LastUsedDate: f.LastUsedDate,
		}
	}

	decKey := ""
	if len(obj.Key) > 0 {
		decKey = string(objectKey.Key)
	}

	return &models.Item{
		Attachments:         decAttachments,
		CollectionIds:       obj.CollectionIds,
		CreationDate:        cloneDate(obj.CreationDate),
		DeletedDate:         cloneDate(obj.DeletedDate),
		Edit:                obj.Edit,
		Favorite:            obj.Favorite,
		Fields:              decFields,
		FolderID:            obj.FolderID,
		ID:                  obj.ID,
		Key:                 decKey,
		Login:               *decLogin,
		Name:                decName,
		Notes:               decNotes,
		Object:              models.ObjectTypeItem,
		OrganizationID:      obj.OrganizationID,
		OrganizationUseTotp: obj.OrganizationUseTotp,
		PasswordHistory:     decPasswordHistory,
		Reprompt:            obj.Reprompt,
		RevisionDate:        cloneDate(obj.RevisionDate),
		SecureNote: models.SecureNote{
			Type: obj.SecureNote.Type,
		},
		Type:         obj.Type,
		ViewPassword: obj.ViewPassword,
	}, nil
}

func decryptItemFields(objFields []models.Field, objectKey symmetrickey.Key) ([]models.Field, error) {
	decFields := make([]models.Field, len(objFields))
	for k, f := range objFields {
		decName, err := decryptStringIfNotEmpty(f.Name, objectKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting field name: %v", err)
		}

		decValue, err := decryptStringIfNotEmpty(f.Value, objectKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting field value: %v", err)
		}

		decFields[k] = models.Field{
			Name:  decName,
			Value: decValue,
			Type:  f.Type,
		}
	}
	return decFields, nil
}

func decryptItemLogin(objLogin models.Login, objectKey symmetrickey.Key) (*models.Login, error) {
	decUsername, err := decryptStringIfNotEmpty(objLogin.Username, objectKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting username: %v", err)
	}

	decPassword, err := decryptStringIfNotEmpty(objLogin.Password, objectKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting password: %v", err)
	}

	decTotp, err := decryptStringIfNotEmpty(objLogin.Totp, objectKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting totp: %v", err)
	}

	decUris := make([]models.LoginURI, len(objLogin.URIs))
	for k, f := range objLogin.URIs {
		decUri, err := decryptStringIfNotEmpty(f.URI, objectKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting uri: %v", err)
		}

		decUris[k] = models.LoginURI{
			URI:   decUri,
			Match: f.Match,
		}
	}

	return &models.Login{
		Username: decUsername,
		Password: decPassword,
		Totp:     decTotp,
		URIs:     decUris,
	}, nil
}

func encryptOrgCollection(_ context.Context, obj models.OrgCollection, secret AccountSecrets, verifyObjectEncryption bool) (*webapi.OrganizationCreationRequest, error) {
	orgKey, err := secret.GetOrganizationKey(obj.OrganizationID)
	if err != nil {
		return nil, err
	}

	collectionName := obj.Name
	if len(collectionName) == 0 {
		return nil, fmt.Errorf("collection name is empty")
	}

	collectionName, err = crypto.EncryptAsString([]byte(collectionName), *orgKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting collection: %w", err)
	}

	encObj := webapi.OrganizationCreationRequest{
		Name:   collectionName,
		Users:  []webapi.OrganizationUser{},
		Groups: []string{},
	}
	if verifyObjectEncryption {
		encObjModified := webapi.Collection{
			Id:             obj.ID,
			Name:           encObj.Name,
			OrganizationId: obj.OrganizationID,
		}
		actualObj, err := decryptOrgCollection(encObjModified, secret)
		if err != nil {
			return nil, fmt.Errorf("error decrypting collection for verification: %w", err)
		}

		err = compareObjects(obj, *actualObj)
		if err != nil {
			return nil, fmt.Errorf("error verifying collection after encryption: %w", err)
		}
	}

	return &encObj, nil

}

func encryptFolder(_ context.Context, obj models.Folder, secret AccountSecrets, verifyObjectEncryption bool) (*webapi.Folder, error) {
	encFolderName, err := encryptAsStringIfNotEmpty(obj.Name, secret.MainKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting folder's name: %w", err)
	}

	encFolder := webapi.Folder{
		Id:           obj.ID,
		Object:       obj.Object,
		RevisionDate: obj.RevisionDate,
		Name:         encFolderName,
	}

	if verifyObjectEncryption {
		actualObj, err := decryptFolder(encFolder, secret)
		if err != nil {
			return nil, fmt.Errorf("error decrypting folder for verification: %w", err)
		}

		err = compareObjects(obj, *actualObj)
		if err != nil {
			return nil, fmt.Errorf("error verifying folder after encryption: %w", err)
		}
	}

	return &encFolder, nil
}

func encryptItem(_ context.Context, obj models.Item, secret AccountSecrets, verifyObjectEncryption bool) (*models.Item, error) {
	objectKey, err := getOrCreateObjectKey(obj)
	if err != nil {
		return nil, err
	}

	mainKey, err := getMainKeyForObject(obj, secret)
	if err != nil {
		return nil, err
	}

	encObjectKey, err := crypto.EncryptAsString(objectKey.Key, *mainKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item object key: %w", err)
	}

	encLogin, err := encryptItemLogin(obj.Login, *objectKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item login: %w", err)
	}

	encFields, err := encryptItemFields(obj.Fields, *objectKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item fields: %w", err)
	}

	encName, err := encryptAsStringIfNotEmpty(obj.Name, *objectKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item name: %w", err)
	}

	encNotes, err := encryptAsStringIfNotEmpty(obj.Notes, *objectKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item notes: %w", err)
	}

	encAttachments := make([]models.Attachment, len(obj.Attachments))
	for k, f := range obj.Attachments {
		encFilename, err := encryptAsStringIfNotEmpty(f.FileName, *objectKey)
		if err != nil {
			return nil, fmt.Errorf("error encrypting attachment filename: %v", err)
		}

		encAttachments[k] = models.Attachment{
			FileName: encFilename,
			ID:       f.ID,
			Key:      f.Key,
			Object:   f.Object,
			Size:     f.Size,
			SizeName: f.SizeName,
			Url:      f.Url,
		}
	}

	encPasswordHistory := make([]models.PasswordHistoryItem, len(obj.PasswordHistory))
	for k, f := range obj.PasswordHistory {
		encPassword, err := encryptAsStringIfNotEmpty(f.Password, *objectKey)
		if err != nil {
			return nil, fmt.Errorf("error encrypting password history: %v", err)
		}

		encPasswordHistory[k] = models.PasswordHistoryItem{
			Password:     encPassword,
			LastUsedDate: f.LastUsedDate,
		}
	}

	encObj := models.Item{
		Attachments:         encAttachments,
		CollectionIds:       obj.CollectionIds,
		CreationDate:        cloneDate(obj.CreationDate),
		DeletedDate:         cloneDate(obj.DeletedDate),
		Edit:                obj.Edit,
		Favorite:            obj.Favorite,
		Fields:              encFields,
		FolderID:            obj.FolderID,
		ID:                  obj.ID,
		Key:                 encObjectKey,
		Login:               *encLogin,
		Name:                encName,
		Notes:               encNotes,
		Object:              obj.Object,
		OrganizationID:      obj.OrganizationID,
		OrganizationUseTotp: obj.OrganizationUseTotp,
		PasswordHistory:     encPasswordHistory,
		Reprompt:            obj.Reprompt,
		RevisionDate:        cloneDate(obj.RevisionDate),
		SecureNote: models.SecureNote{
			Type: obj.SecureNote.Type,
		},
		Type:         obj.Type,
		ViewPassword: obj.ViewPassword,
	}

	if verifyObjectEncryption {
		actualObj, err := decryptItem(encObj, secret)
		if err != nil {
			return nil, fmt.Errorf("error decrypting item for verification: %w", err)
		}

		actualObj.Key = ""

		err = compareObjects(obj, *actualObj)
		if err != nil {
			return nil, fmt.Errorf("error verifying item after encryption: %w", err)
		}
	}

	return &encObj, nil
}

func encryptItemFields(objFields []models.Field, objectKey symmetrickey.Key) ([]models.Field, error) {
	encObjFields := make([]models.Field, len(objFields))
	for k, v := range objFields {
		encryptedName, err := encryptAsStringIfNotEmpty(v.Name, objectKey)
		if err != nil {
			return nil, fmt.Errorf("error encrypting field's name: %w", err)
		}

		encryptedValue, err := encryptAsStringIfNotEmpty(v.Value, objectKey)
		if err != nil {
			return nil, fmt.Errorf("error encrypting field's value: %w", err)
		}

		encObjFields[k] = models.Field{
			Name:  encryptedName,
			Value: encryptedValue,
			Type:  objFields[k].Type,
		}

	}
	return encObjFields, nil
}

func encryptItemLogin(objLogin models.Login, objectKey symmetrickey.Key) (*models.Login, error) {
	encUsername, err := encryptAsStringIfNotEmpty(objLogin.Username, objectKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item login username: %w", err)
	}

	encPassword, err := encryptAsStringIfNotEmpty(objLogin.Password, objectKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item login password: %w", err)
	}

	encTotp, err := encryptAsStringIfNotEmpty(objLogin.Totp, objectKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting item login totp: %w", err)
	}

	encUris := make([]models.LoginURI, len(objLogin.URIs))
	for k, f := range objLogin.URIs {
		encUri := models.LoginURI{
			URI:   f.URI,
			Match: f.Match,
		}

		encUri.URI, err = encryptAsStringIfNotEmpty(f.URI, objectKey)
		if err != nil {
			return nil, fmt.Errorf("error encrypting uri: %w", err)
		}
		encUris[k] = encUri
	}

	return &models.Login{
		Username: encUsername,
		Password: encPassword,
		Totp:     encTotp,
		URIs:     encUris,
	}, nil
}

func (v *baseVault) getOrDefaultObjectKey(obj models.Item) (*symmetrickey.Key, error) {
	if len(obj.Key) == 0 {
		return &v.loginAccount.Secrets.MainKey, nil
	}

	return symmetrickey.NewFromRawBytesWithEncryptionType([]byte(obj.Key), symmetrickey.AesCbc256_HmacSha256_B64)
}

func getMainKeyForObject(obj models.Item, secret AccountSecrets) (*symmetrickey.Key, error) {
	if len(obj.OrganizationID) == 0 {
		return &secret.MainKey, nil
	}

	return secret.GetOrganizationKey(obj.OrganizationID)
}

func getObjectKey(obj models.Item, secret AccountSecrets) (*symmetrickey.Key, error) {
	objectKey, err := getMainKeyForObject(obj, secret)
	if err != nil {
		return nil, err
	}

	if len(obj.Key) > 0 {
		decryptedObjectKey, err := decryptStringAsKey(obj.Key, *objectKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting object key: %w", err)
		}
		objectKey = decryptedObjectKey
	}
	return objectKey, nil
}

func getOrCreateObjectKey(obj models.Item) (*symmetrickey.Key, error) {
	var objectKeyBytes []byte
	if len(obj.Key) > 0 {
		objectKeyBytes = []byte(obj.Key)
		return symmetrickey.NewFromRawBytesWithEncryptionType(objectKeyBytes, symmetrickey.AesCbc256_HmacSha256_B64)
	} else {
		return keybuilder.CreateObjectKey()
	}
}

func objKey(obj any) string {
	switch itemObj := any(obj).(type) {
	case models.Item:
		return fmt.Sprintf("%s___%s", models.ObjectTypeItem, itemObj.ID)
	case models.Folder:
		return fmt.Sprintf("%s___%s", models.ObjectTypeFolder, itemObj.ID)
	case models.Organization:
		return fmt.Sprintf("%s___%s", models.ObjectTypeOrganization, itemObj.ID)
	case models.OrgCollection:
		return fmt.Sprintf("%s___%s", models.ObjectTypeOrgCollection, itemObj.ID)
	}
	panic(fmt.Sprintf("BUG: objKey, unsupported object type: %T", obj))
}

func objMatchFilter[T any](ctx context.Context, rawObj T, filters bitwarden.ListObjectsFilterOptions, objType models.ObjectType) bool {
	switch obj := any(rawObj).(type) {
	case models.Item:
		if obj.DeletedDate != nil {
			tflog.Trace(ctx, "Ignoring deleted object in search results", map[string]interface{}{"object_id": obj.ID})
			return false
		}

		if obj.Object != objType {
			return false
		}

		if len(filters.OrganizationFilter) > 0 && obj.OrganizationID != filters.OrganizationFilter {
			return false
		}
		if filters.ItemType > 0 && obj.Type != filters.ItemType {
			return false
		}
		if len(filters.FolderFilter) > 0 && obj.FolderID != filters.FolderFilter {
			return false
		}

		if len(filters.CollectionFilter) > 0 {
			matchCollection := false
			for _, h := range obj.CollectionIds {
				if h == filters.CollectionFilter {
					matchCollection = true
				}
			}
			if !matchCollection {
				return false
			}
		}

		if len(filters.UrlFilter) > 0 {
			matchUrl := false
			for _, u := range obj.Login.URIs {
				if u.Match == nil {
					// When selecting 'default' match in the CLI, it results in a
					// 'nil' match which we default to 'base_domain' here.
					u.Match = models.URIMatchBaseDomain.ToPointer()
				}

				matched, err := urlsMatch(u, filters.UrlFilter)
				if err != nil {
					tflog.Trace(ctx, "Error matching URL", map[string]interface{}{"object_id": obj.ID, "url": u.URI, "error": err})
					continue
				}
				if matched {
					matchUrl = true
				}
			}
			if !matchUrl {
				return false
			}
		}
	case models.OrgCollection:
		if obj.Object != objType {
			return false
		}

		if len(filters.OrganizationFilter) > 0 && obj.OrganizationID != filters.OrganizationFilter {
			return false
		}
	case models.Folder:
		if obj.Object != objType {
			return false
		}
	case models.Organization:
		if obj.Object != objType {
			return false
		}
	}

	if len(filters.SearchFilter) > 0 {
		foundSomething := false
		switch obj := any(rawObj).(type) {
		case models.Item:
			if strings.Contains(obj.Name, filters.SearchFilter) {
				foundSomething = true
			}
			if strings.Contains(obj.Login.Username, filters.SearchFilter) {
				foundSomething = true
			}

			if strings.Contains(obj.Notes, filters.SearchFilter) {
				foundSomething = true
			}
		case models.Folder:
			if strings.Contains(obj.Name, filters.SearchFilter) {
				foundSomething = true
			}
		case models.OrgCollection:
			if strings.Contains(obj.Name, filters.SearchFilter) {
				foundSomething = true
			}
		case models.Organization:
			if strings.Contains(obj.Name, filters.SearchFilter) {
				foundSomething = true
			}
		}

		if !foundSomething {
			return false
		}
	}
	return len(filters.SearchFilter) > 0
}

func urlsMatch(u models.LoginURI, searchedUrl string) (bool, error) {
	if u.Match == nil {
		return false, nil
	}

	switch *u.Match {
	case models.URIMatchBaseDomain:
		// TODO: Support equivalent domains
		return domainsMatch(u.URI, searchedUrl)
	case models.URIMatchHost:
		return matchHost(u.URI, searchedUrl)
	case models.URIMatchStartWith:
		return strings.HasPrefix(searchedUrl, u.URI), nil
	case models.URIMatchExact:
		return searchedUrl == u.URI, nil
	case models.URIMatchRegExp:
		matched, err := regexp.MatchString(u.URI, searchedUrl)
		if err != nil {
			return false, fmt.Errorf("error matching regex: %w", err)
		}
		return matched, nil
	case models.URIMatchNever:
		return false, nil
	default:
		return false, fmt.Errorf("unsupported URIMatch: %d", *u.Match)
	}
}

func domainsMatch(url1, url2 string) (bool, error) {
	parsedUrl1, err := url.Parse(url1)
	if err != nil {
		return false, fmt.Errorf("error parsing url1: %w", err)
	}

	parsedUrl1Domain, err := publicsuffix.EffectiveTLDPlusOne(parsedUrl1.Host)
	if err != nil {
		return false, fmt.Errorf("error getting url1 TLD+1: %w", err)
	}

	parsedUrl2, err := url.Parse(url2)
	if err != nil {
		return false, fmt.Errorf("error parsing url2: %w", err)
	}

	parsedUrl2Domain, err := publicsuffix.EffectiveTLDPlusOne(parsedUrl2.Host)
	if err != nil {
		return false, fmt.Errorf("error getting url2 TLD+1: %w", err)
	}

	return len(parsedUrl1Domain) > 0 && parsedUrl1Domain == parsedUrl2Domain, nil
}

func matchHost(url1, url2 string) (bool, error) {
	parsedUrl1, err := url.Parse(url1)
	if err != nil {
		return false, fmt.Errorf("error parsing url1: %w", err)
	}

	parsedUrl2, err := url.Parse(url2)
	if err != nil {
		return false, fmt.Errorf("error parsing url2: %w", err)
	}
	return len(parsedUrl1.Host) > 0 && parsedUrl1.Host == parsedUrl2.Host, nil
}

func compareObjects[T any](actual, expected T) error {
	actualJson, err := json.Marshal(actual)
	if err != nil {
		err := fmt.Errorf("error marshalling actual object while comparing: %w", err)
		if panicOnEncryptionErrors {
			panic(err)
		}
		return err
	}
	expectedJson, err := json.Marshal(expected)
	if err != nil {
		err := fmt.Errorf("error marshalling expected object while comparing: %w", err)
		if panicOnEncryptionErrors {
			panic(err)
		}
		return err
	}

	if !bytes.Equal(actualJson, expectedJson) {
		err := fmt.Errorf("object comparison failed")
		fmt.Printf("Expected: %s\n", string(actualJson))
		fmt.Printf("Actual: %s\n", string(expectedJson))
		if panicOnEncryptionErrors {
			panic(err)
		}
		return err
	}
	return nil
}

func cloneDate(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	clone := *t
	return &clone
}
