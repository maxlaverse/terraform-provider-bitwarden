package embedded

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
)

var (
	panicOnEncryptionErrors = false
)

type BaseVault interface {
	GetObject(ctx context.Context, obj models.Object) (*models.Object, error)
	ListObjects(ctx context.Context, objType models.ObjectType, options ...bitwarden.ListObjectsOption) ([]models.Object, error)
}

type baseVault struct {
	loginAccount Account
	objectStore  map[string]models.Object

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

func (v *baseVault) GetObject(ctx context.Context, obj models.Object) (*models.Object, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	return v.getObject(ctx, obj)
}

func (v *baseVault) getObject(_ context.Context, obj models.Object) (*models.Object, error) {
	if v.objectStore == nil {
		return nil, models.ErrVaultLocked
	}

	storedObj, ok := v.objectStore[objKey(obj)]
	if !ok || obj.DeletedDate != nil {
		return nil, models.ErrObjectNotFound
	}

	return &storedObj, nil
}

func (v *baseVault) ListObjects(ctx context.Context, objType models.ObjectType, options ...bitwarden.ListObjectsOption) ([]models.Object, error) {
	v.vaultOperationMutex.RLock()
	defer v.vaultOperationMutex.RUnlock()

	if v.objectStore == nil {
		return nil, models.ErrVaultLocked
	}

	filter := bitwarden.ListObjectsOptionsToFilterOptions(options...)
	if !filter.IsValid() {
		return nil, fmt.Errorf("invalid filter options")
	}

	foundObjects := []models.Object{}
	for _, obj := range v.objectStore {
		if obj.Object != objType {
			continue
		}

		if obj.DeletedDate != nil {
			tflog.Trace(ctx, "Ignoring deleted object in search results", map[string]interface{}{"object_id": obj.ID})
			continue
		}

		if !objMatchFilter(obj, filter) {
			continue
		}

		foundObjects = append(foundObjects, obj)
	}

	return foundObjects, nil
}

func (v *baseVault) clearObjectStore(ctx context.Context) {
	if v.objectStore != nil {
		tflog.Trace(ctx, "Clearing object store")
	}
	v.objectStore = make(map[string]models.Object)
}

func (v *baseVault) deleteObjectFromStore(ctx context.Context, obj models.Object) {
	tflog.Trace(ctx, "Deleting object from store", map[string]interface{}{"object_id": obj.ID, "object_name": obj.Name, "object_folder_id": obj.FolderID})
	delete(v.objectStore, objKey(obj))
}

func (v *baseVault) encryptFolder(_ context.Context, obj models.Object, secret AccountSecrets) (*webapi.Folder, error) {
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

	if v.verifyObjectEncryption {
		objForVerification, err := decryptFolder(encFolder, secret)
		if err != nil {
			return nil, fmt.Errorf("error decrypting folder for verification: %w", err)
		}

		err = compareObjects(obj, *objForVerification)
		if err != nil {
			return nil, fmt.Errorf("error verifying folder after encryption: %w", err)
		}
	}

	return &encFolder, nil
}

func (v *baseVault) objectsLoaded() bool {
	return v.objectStore != nil
}

func (v *baseVault) storeObject(ctx context.Context, obj models.Object) {
	tflog.Trace(ctx, "Storing new object", map[string]interface{}{"object_id": obj.ID, "object_name": obj.Name, "object_folder_id": obj.FolderID})
	v.objectStore[objKey(obj)] = obj
}

func (v *baseVault) storeObjects(ctx context.Context, objs []models.Object) {
	for _, obj := range objs {
		v.storeObject(ctx, obj)
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

func decryptCollection(obj webapi.Collection, secret AccountSecrets) (*models.Object, error) {
	orgKey, err := secret.GetOrganizationKey(obj.OrganizationId)
	if err != nil {
		return nil, err
	}

	decName, err := decryptStringIfNotEmpty(obj.Name, *orgKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting collection name: %w", err)
	}

	return &models.Object{
		ID:             obj.Id,
		Name:           decName,
		Object:         models.ObjectTypeOrgCollection,
		OrganizationID: obj.OrganizationId,
	}, nil
}

func decryptFolder(obj webapi.Folder, secret AccountSecrets) (*models.Object, error) {
	objName, err := decryptStringIfNotEmpty(obj.Name, secret.MainKey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting folder name: %w", err)
	}

	return &models.Object{
		ID:           obj.Id,
		Name:         objName,
		Object:       models.ObjectTypeFolder,
		RevisionDate: cloneDate(obj.RevisionDate),
	}, nil
}

func decryptItem(obj models.Object, secret AccountSecrets) (*models.Object, error) {
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

	return &models.Object{
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
		OrganizationID:      obj.OrganizationID,
		OrganizationUseTotp: obj.OrganizationUseTotp,
		Object:              models.ObjectTypeItem,
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

func encryptCollection(obj models.Object, secret AccountSecrets, verifyObjectEncryption bool) (*webapi.OrganizationCreationRequest, error) {
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
		objForVerification, err := decryptCollection(encObjModified, secret)
		if err != nil {
			return nil, fmt.Errorf("error decrypting collection for verification: %w", err)
		}

		objForVerification.Key = ""

		err = compareObjects(obj, *objForVerification)
		if err != nil {
			return nil, fmt.Errorf("error verifying collection after encryption: %w", err)
		}
	}

	return &encObj, nil

}

func encryptItem(_ context.Context, obj models.Object, secret AccountSecrets, verifyObjectEncryption bool) (*models.Object, error) {
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

	encObj := models.Object{
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
		OrganizationID:      obj.OrganizationID,
		OrganizationUseTotp: obj.OrganizationUseTotp,
		Object:              obj.Object,
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
		objForVerification, err := decryptItem(encObj, secret)
		if err != nil {
			return nil, fmt.Errorf("error decrypting item for verification: %w", err)
		}

		objForVerification.Key = ""

		err = compareObjects(obj, *objForVerification)
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

func (v *baseVault) getOrDefaultObjectKey(obj models.Object) (*symmetrickey.Key, error) {
	if len(obj.Key) == 0 {
		return &v.loginAccount.Secrets.MainKey, nil
	}

	return symmetrickey.NewFromRawBytesWithEncryptionType([]byte(obj.Key), symmetrickey.AesCbc256_HmacSha256_B64)
}

func getMainKeyForObject(obj models.Object, secret AccountSecrets) (*symmetrickey.Key, error) {
	if len(obj.OrganizationID) == 0 {
		return &secret.MainKey, nil
	}

	return secret.GetOrganizationKey(obj.OrganizationID)
}

func getObjectKey(obj models.Object, secret AccountSecrets) (*symmetrickey.Key, error) {
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

func getOrCreateObjectKey(obj models.Object) (*symmetrickey.Key, error) {
	var objectKeyBytes []byte
	if len(obj.Key) > 0 {
		objectKeyBytes = []byte(obj.Key)
		return symmetrickey.NewFromRawBytesWithEncryptionType(objectKeyBytes, symmetrickey.AesCbc256_HmacSha256_B64)
	} else {
		return keybuilder.CreateObjectKey()
	}
}

func objKey(obj models.Object) string {
	return fmt.Sprintf("%s___%s", obj.Object, obj.ID)
}

func objMatchFilter(obj models.Object, filters bitwarden.ListObjectsFilterOptions) bool {
	if len(filters.OrganizationFilter) > 0 && obj.OrganizationID != filters.OrganizationFilter {
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

	if len(filters.SearchFilter) > 0 {
		foundSomething := false
		if strings.Contains(obj.Name, filters.SearchFilter) {
			foundSomething = true
		}
		if strings.Contains(obj.Login.Username, filters.SearchFilter) {
			foundSomething = true
		}
		if strings.Contains(obj.Notes, filters.SearchFilter) {
			foundSomething = true
		}
		if !foundSomething {
			return false
		}
	}
	return len(filters.SearchFilter) > 0
}

func compareObjects(obj1, obj2 models.Object) error {
	out1, err := json.Marshal(obj1)
	if err != nil {
		err := fmt.Errorf("error marshalling obj1 while comparing: %w", err)
		if panicOnEncryptionErrors {
			panic(err)
		}
		return err
	}
	out2, err := json.Marshal(obj2)
	if err != nil {
		err := fmt.Errorf("error marshalling obj2 while comparing: %w", err)
		if panicOnEncryptionErrors {
			panic(err)
		}
		return err
	}

	if !bytes.Equal(out1, out2) {
		err := fmt.Errorf("object comparison failed")
		fmt.Printf("Object1: %s\n", string(out1))
		fmt.Printf("Object2: %s\n", string(out2))
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
