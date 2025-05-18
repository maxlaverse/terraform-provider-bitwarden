//go:build offline

package embedded

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/stretchr/testify/assert"
)

const (
	orgKey  = "4.JW3mktbL7vpTVRweZdQBAirJuAEhSRn37zcXZjDjI47weFKkeZkvPxZWCqFYC/P5qCJwEYMbv7lTETkWDg6paevVfhJ35buGcTQdEbQxAJebzPahEcUstj11l4Y9T5RaDiAJR8+drrGJ3fKV3v3hymKz2o9fUfK1epuLFll2nnWSOjCcuRe/+zz5VwIVx4WJAPJHmiS6eofbj/DTIQCzG4JkR0UzT66ouLcgmPL1nGOqVI7KxRpL5yVj75UkjniHkWAcB7lfAxWXw2GhDJ/2L685uA3820ItTbxjCwLQOvjBttgrbURmkeP9BD+KkO4V6vb8bbTWNSvggXKk2h1CMw=="
	orgUuid = "81cc1652-dc80-472d-909f-9539d057068b"
)

func TestCompareObjectsSimpleKeys(t *testing.T) {
	obj1 := models.Item{
		Name: "test",
	}
	obj2 := models.Item{
		Name: "test",
	}
	obj3 := models.Item{
		Name: "test1",
	}
	assert.NoError(t, compareObjects(context.Background(), obj1, obj2))
	assert.EqualError(t, compareObjects(context.Background(), obj1, obj3), "different keys at [/name]")
	assert.NoError(t, compareObjects(context.Background(), obj1, obj3, "/name"))
}

func TestCompareObjectsArrays(t *testing.T) {
	obj1 := models.Item{
		Name:          "test",
		CollectionIds: []string{"1"},
	}
	obj2 := models.Item{
		Name:          "test",
		CollectionIds: []string{"2"},
	}
	obj3 := models.Item{
		Name:          "test1",
		CollectionIds: nil,
	}
	obj4 := models.Item{
		Name:          "test1",
		CollectionIds: []string{"1", "2"},
	}
	obj5 := models.Item{
		Name:          "test1",
		CollectionIds: []string{"2", "1"},
	}
	assert.NoError(t, compareObjects(context.Background(), obj1, obj1))
	assert.EqualError(t, compareObjects(context.Background(), obj1, obj2), "different keys at [/collectionIds/0]")
	assert.NoError(t, compareObjects(context.Background(), obj3, obj4, "/collectionIds"))
	assert.NoError(t, compareObjects(context.Background(), obj4, obj5))
}

func TestCompareNestedObjects(t *testing.T) {
	obj1 := models.Item{
		Attachments: []models.Attachment{
			{
				Url: "https://example.com/test1",
				ID:  "1",
			},
		},
	}
	obj2 := models.Item{
		Attachments: []models.Attachment{
			{
				Url: "https://example.com/test2",
				ID:  "1",
			},
		},
	}
	obj3 := models.Item{
		Attachments: []models.Attachment{
			{
				Url: "https://example.com/test1",
				ID:  "2",
			},
		},
	}
	assert.NoError(t, compareObjects(context.Background(), obj1, obj2, "/attachments/*/url"))
	assert.Error(t, compareObjects(context.Background(), obj1, obj3, "/attachments/*/url"))
}

func TestMatchUrl(t *testing.T) {
	testCases := []struct {
		loginUri        models.LoginURI
		matchingUrls    []string
		notMatchingUrls []string
	}{
		{
			loginUri:        models.LoginURI{URI: "https://sub.mydomain1.com", Match: models.URIMatchBaseDomain.ToPointer()},
			matchingUrls:    []string{"https://mydomain1.com", "https://else.mydomain1.com"},
			notMatchingUrls: []string{"https://mydomain1bis.com"},
		},
		{
			loginUri:        models.LoginURI{URI: "https://mydomain2.com", Match: models.URIMatchHost.ToPointer()},
			matchingUrls:    []string{"https://mydomain2.com"},
			notMatchingUrls: []string{"https://mydomain2bis.com", "https://test.mydomain2.com"},
		},
		{
			loginUri:        models.LoginURI{URI: "https://mydomain3.com/product", Match: models.URIMatchStartWith.ToPointer()},
			matchingUrls:    []string{"https://mydomain3.com/product/page"},
			notMatchingUrls: []string{"https://mydomain3.com/otherproduct/product"},
		},
		{
			loginUri:        models.LoginURI{URI: "https://mydomain4.com/page", Match: models.URIMatchExact.ToPointer()},
			matchingUrls:    []string{"https://mydomain4.com/page"},
			notMatchingUrls: []string{"https://mydomain4.com/page-other"},
		},
		{
			loginUri:        models.LoginURI{URI: "https://mydomain5.com/([a-z]+)/test", Match: models.URIMatchRegExp.ToPointer()},
			matchingUrls:    []string{"https://mydomain5.com/mypage/test"},
			notMatchingUrls: []string{"https://mydomain5.com/mypage2/test"},
		},
		{
			loginUri:        models.LoginURI{URI: "https://mydomain6.com", Match: models.URIMatchNever.ToPointer()},
			notMatchingUrls: []string{"https://mydomain6.com"},
		},
	}

	for _, test := range testCases {
		for _, m := range test.matchingUrls {
			t.Run(fmt.Sprintf("%s ? %s", test.loginUri.URI, m), func(t *testing.T) {
				match, err := urlsMatch(test.loginUri, m)
				assert.NoError(t, err)
				assert.Equal(t, true, match)
			})
		}
		for _, m := range test.notMatchingUrls {
			t.Run(fmt.Sprintf("%s ? %s", test.loginUri.URI, m), func(t *testing.T) {
				match, err := urlsMatch(test.loginUri, m)
				assert.NoError(t, err)
				assert.Equal(t, false, match)
			})
		}
	}
}

func TestDecryptAccountSecretPbkdf2(t *testing.T) {
	accountSecrets, err := decryptAccountSecrets(AccountPbkdf2, TestPassword)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "NJ0uDK79BMZShPanVpmJM8efx5VFlij9wzf92Sys59o=", accountSecrets.MasterPasswordHash)

	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(accountSecrets.RSAPrivateKey),
		},
	)

	assert.Equal(t, RsaPrivateKey, strings.Replace(string(pemdata), "\\n", "\n", -1))
	assert.Contains(t, accountSecrets.MainKey.Summary(), EncryptionKey)
}

func TestDecryptAccountSecretArgon2(t *testing.T) {
	accountSecrets, err := decryptAccountSecrets(AccountArgon2, TestPassword)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "RMnX3/HmNS18BK54HrfBB4EFglSesU5Z7+fV6v8FaKY=", accountSecrets.MasterPasswordHash)

	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(accountSecrets.RSAPrivateKey),
		},
	)

	assert.Equal(t, RsaPrivateKey, strings.Replace(string(pemdata), "\\n", "\n", -1))
	assert.Contains(t, accountSecrets.MainKey.Summary(), EncryptionKey)
}

func TestDecryptAccountSecretWrongPassword(t *testing.T) {
	_, err := decryptAccountSecrets(AccountPbkdf2, "wrong-password")
	assert.Error(t, err, "decryption should fail with wrong password")
	assert.ErrorIs(t, err, models.ErrWrongMasterPassword)
}

func TestEncryptOrgCollection(t *testing.T) {
	accountSecrets := computeTestAccountSecrets(t)

	orgToEncrypt := testFullyFilledOrgCollection()
	newFolder, err := encryptOrgCollection(context.Background(), orgToEncrypt, *accountSecrets, true)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, "sensitive-name", orgToEncrypt.Name)
	assertEncryptedValueOf(t, "sensitive-name", newFolder.Name, accountSecrets.OrganizationSecrets[orgUuid].Key)

	newOut, err := json.Marshal(newFolder)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotContains(t, string(newOut), "sensitive")
}

func TestEncryptFolder(t *testing.T) {
	accountSecrets := computeTestAccountSecrets(t)

	folderToEncrypt := testFullyFilledFolder()
	newFolder, err := encryptFolder(context.Background(), folderToEncrypt, *accountSecrets, true)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, "sensitive-name", folderToEncrypt.Name)
	assertEncryptedValueOf(t, "sensitive-name", newFolder.Name, accountSecrets.MainKey)

	newOut, err := json.Marshal(newFolder)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotContains(t, string(newOut), "sensitive")
}

func TestEncryptItem(t *testing.T) {
	accountSecrets := computeTestAccountSecrets(t)

	objectToEncrypt := testFullyFilledItem()
	newObj, err := encryptItem(context.Background(), objectToEncrypt, *accountSecrets, true)
	if !assert.Nil(t, err) {
		return
	}

	r, e := getObjectKey(*newObj, *accountSecrets)
	assert.Nil(t, e)

	assert.Equal(t, []string{"test-collection-id"}, objectToEncrypt.CollectionIds)
	assert.Equal(t, []string{"test-collection-id"}, newObj.CollectionIds)

	if assert.NotNil(t, *newObj.CreationDate) {
		assert.Equal(t, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), *objectToEncrypt.CreationDate)
		assert.Equal(t, time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC), *newObj.CreationDate)
	}

	if assert.NotNil(t, *newObj.DeletedDate) {
		assert.Equal(t, time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC), *objectToEncrypt.DeletedDate)
		assert.Equal(t, time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC), *newObj.DeletedDate)
	}

	assert.Equal(t, true, objectToEncrypt.Edit)
	assert.Equal(t, true, newObj.Edit)

	assert.Equal(t, true, objectToEncrypt.Favorite)
	assert.Equal(t, true, newObj.Favorite)

	if assert.Len(t, newObj.Attachments, 1) {
		assert.Equal(t, "public-id", objectToEncrypt.Attachments[0].ID)
		assert.Equal(t, "public-id", newObj.Attachments[0].ID)

		assert.Equal(t, "sensitive-filename", objectToEncrypt.Attachments[0].FileName)
		assertEncryptedValueOf(t, "sensitive-filename", newObj.Attachments[0].FileName, *r)

		assert.Equal(t, "public-size", objectToEncrypt.Attachments[0].Size)
		assert.Equal(t, "public-size", newObj.Attachments[0].Size)

		assert.Equal(t, "public-size-name", objectToEncrypt.Attachments[0].SizeName)
		assert.Equal(t, "public-size-name", newObj.Attachments[0].SizeName)

		assert.Equal(t, "public-url", objectToEncrypt.Attachments[0].Url)
		assert.Equal(t, "public-url", newObj.Attachments[0].Url)

		assert.Equal(t, "already-encrypted-key", objectToEncrypt.Attachments[0].Key)
		assert.Equal(t, "already-encrypted-key", newObj.Attachments[0].Key)
	}

	if assert.Len(t, newObj.Fields, 4) {
		assert.Equal(t, "sensitive-boolfield-name", objectToEncrypt.Fields[0].Name)
		assertEncryptedValueOf(t, "sensitive-boolfield-name", newObj.Fields[0].Name, *r)
		assert.Equal(t, "sensitive-true", objectToEncrypt.Fields[0].Value)
		assertEncryptedValueOf(t, "sensitive-true", newObj.Fields[0].Value, *r)
		assert.Equal(t, models.FieldTypeBoolean, objectToEncrypt.Fields[0].Type)
		assert.Equal(t, models.FieldTypeBoolean, newObj.Fields[0].Type)

		assert.Equal(t, "sensitive-hiddenfield-name", objectToEncrypt.Fields[1].Name)
		assertEncryptedValueOf(t, "sensitive-hiddenfield-name", newObj.Fields[1].Name, *r)
		assert.Equal(t, "sensitive-hiddenfield-value", objectToEncrypt.Fields[1].Value)
		assertEncryptedValueOf(t, "sensitive-hiddenfield-value", newObj.Fields[1].Value, *r)
		assert.Equal(t, models.FieldTypeHidden, objectToEncrypt.Fields[1].Type)
		assert.Equal(t, models.FieldTypeHidden, newObj.Fields[1].Type)

		assert.Equal(t, "sensitive-linkedfield-name", objectToEncrypt.Fields[2].Name)
		assertEncryptedValueOf(t, "sensitive-linkedfield-name", newObj.Fields[2].Name, *r)
		assert.Equal(t, "sensitive-linkedfield-value", objectToEncrypt.Fields[2].Value)
		assertEncryptedValueOf(t, "sensitive-linkedfield-value", newObj.Fields[2].Value, *r)
		assert.Equal(t, models.FieldTypeLinked, objectToEncrypt.Fields[2].Type)
		assert.Equal(t, models.FieldTypeLinked, newObj.Fields[2].Type)

		assert.Equal(t, "sensitive-textfield-name", objectToEncrypt.Fields[3].Name)
		assertEncryptedValueOf(t, "sensitive-textfield-name", newObj.Fields[3].Name, *r)
		assert.Equal(t, "sensitive-textfield-value", objectToEncrypt.Fields[3].Value)
		assertEncryptedValueOf(t, "sensitive-textfield-value", newObj.Fields[3].Value, *r)
		assert.Equal(t, models.FieldTypeText, objectToEncrypt.Fields[3].Type)
		assert.Equal(t, models.FieldTypeText, newObj.Fields[3].Type)
	}

	assert.Equal(t, "test-folder-id", objectToEncrypt.FolderID)
	assert.Equal(t, "test-folder-id", newObj.FolderID)

	assert.Equal(t, "test-id", objectToEncrypt.ID)
	assert.Equal(t, "test-id", newObj.ID)

	assert.Equal(t, "sensitive-username", objectToEncrypt.Login.Username)
	assertEncryptedValueOf(t, "sensitive-username", newObj.Login.Username, *r)

	assert.Equal(t, "sensitive-password", objectToEncrypt.Login.Password)
	assertEncryptedValueOf(t, "sensitive-password", newObj.Login.Password, *r)

	assert.Equal(t, "sensitive-totp", objectToEncrypt.Login.Totp)
	assertEncryptedValueOf(t, "sensitive-totp", newObj.Login.Totp, *r)

	if assert.Len(t, newObj.Login.URIs, 5) {
		assert.Equal(t, "sensitive-uri-basedomain", objectToEncrypt.Login.URIs[0].URI)
		assertEncryptedValueOf(t, "sensitive-uri-basedomain", newObj.Login.URIs[0].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[0].Match) {
			assert.Equal(t, models.URIMatchBaseDomain, *objectToEncrypt.Login.URIs[0].Match)
			assert.Equal(t, models.URIMatchBaseDomain, *newObj.Login.URIs[0].Match)
		}

		assert.Equal(t, "sensitive-uri-exact", objectToEncrypt.Login.URIs[1].URI)
		assertEncryptedValueOf(t, "sensitive-uri-exact", newObj.Login.URIs[1].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[1].Match) {
			assert.Equal(t, models.URIMatchExact, *objectToEncrypt.Login.URIs[1].Match)
			assert.Equal(t, models.URIMatchExact, *newObj.Login.URIs[1].Match)
		}

		assert.Equal(t, "sensitive-uri-never", objectToEncrypt.Login.URIs[2].URI)
		assertEncryptedValueOf(t, "sensitive-uri-never", newObj.Login.URIs[2].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[2].Match) {
			assert.Equal(t, models.URIMatchNever, *objectToEncrypt.Login.URIs[2].Match)
			assert.Equal(t, models.URIMatchNever, *newObj.Login.URIs[2].Match)
		}

		assert.Equal(t, "sensitive-uri-regexp", objectToEncrypt.Login.URIs[3].URI)
		assertEncryptedValueOf(t, "sensitive-uri-regexp", newObj.Login.URIs[3].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[3].Match) {
			assert.Equal(t, models.URIMatchRegExp, *objectToEncrypt.Login.URIs[3].Match)
			assert.Equal(t, models.URIMatchRegExp, *newObj.Login.URIs[3].Match)
		}

		assert.Equal(t, "sensitive-uri-startwith", objectToEncrypt.Login.URIs[4].URI)
		assertEncryptedValueOf(t, "sensitive-uri-startwith", newObj.Login.URIs[4].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[4].Match) {
			assert.Equal(t, models.URIMatchStartWith, *objectToEncrypt.Login.URIs[4].Match)
			assert.Equal(t, models.URIMatchStartWith, *newObj.Login.URIs[4].Match)
		}
	}

	assert.Equal(t, "sensitive-name", objectToEncrypt.Name)
	assertEncryptedValueOf(t, "sensitive-name", newObj.Name, *r)

	assert.Equal(t, "sensitive-notes", objectToEncrypt.Notes)
	assertEncryptedValueOf(t, "sensitive-notes", newObj.Notes, *r)

	assert.Equal(t, models.ObjectTypeItem, objectToEncrypt.Object)
	assert.Equal(t, models.ObjectTypeItem, newObj.Object)

	assert.Equal(t, "81cc1652-dc80-472d-909f-9539d057068b", objectToEncrypt.OrganizationID)
	assert.Equal(t, "81cc1652-dc80-472d-909f-9539d057068b", newObj.OrganizationID)

	assert.Equal(t, true, objectToEncrypt.OrganizationUseTotp)
	assert.Equal(t, true, newObj.OrganizationUseTotp)

	if assert.Len(t, newObj.PasswordHistory, 1) {
		assert.Equal(t, "sensitive-password", objectToEncrypt.PasswordHistory[0].Password)
		assertEncryptedValueOf(t, "sensitive-password", newObj.PasswordHistory[0].Password, *r)

		if assert.NotNil(t, newObj.PasswordHistory[0].LastUsedDate) {
			assert.Equal(t, time.Date(2021, time.February, 1, 0, 0, 0, 0, time.UTC), *objectToEncrypt.PasswordHistory[0].LastUsedDate)
			assert.Equal(t, time.Date(2021, time.February, 1, 0, 0, 0, 0, time.UTC), *newObj.PasswordHistory[0].LastUsedDate)
		}
	}

	assert.Equal(t, 1, objectToEncrypt.Reprompt)
	assert.Equal(t, 1, newObj.Reprompt)

	if assert.NotNil(t, *newObj.RevisionDate) {
		assert.Equal(t, time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC), *objectToEncrypt.RevisionDate)
		assert.Equal(t, time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC), *newObj.RevisionDate)
	}

	assert.Equal(t, 3, objectToEncrypt.SecureNote.Type)
	assert.Equal(t, 3, newObj.SecureNote.Type)

	assert.Equal(t, models.ItemTypeLogin, objectToEncrypt.Type)
	assert.Equal(t, models.ItemTypeLogin, newObj.Type)

	assert.Equal(t, true, objectToEncrypt.ViewPassword)
	assert.Equal(t, true, newObj.ViewPassword)

	newOut, err := json.Marshal(newObj)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotContains(t, string(newOut), "sensitive")
}

func assertEncryptedValueOf(t *testing.T, expected, value string, k symmetrickey.Key) {
	out, err := decryptStringAsBytes(value, k)
	assert.Nil(t, err)
	assert.Equal(t, expected, string(out))
}

func computeTestAccountSecrets(t *testing.T) *AccountSecrets {
	accountSecrets, err := decryptAccountSecrets(AccountPbkdf2, TestPassword)
	if err != nil {
		t.Fatal(err)
	}

	key, err := decryptOrganizationKey(orgKey, accountSecrets.RSAPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	accountSecrets.OrganizationSecrets = map[string]OrganizationSecret{
		orgUuid: {
			OrganizationUUID: orgUuid,
			Key:              *key,
		},
	}
	return accountSecrets
}

/*
 * Test data: all fields meant to be encrypted are prefixed with "sensitive-"
 *            in order to detect eventual leaks while testing.
 */
func testFullyFilledOrgCollection() models.OrgCollection {
	obj := models.OrgCollection{
		Name:           "sensitive-name",
		Object:         models.ObjectTypeOrgCollection,
		OrganizationID: orgUuid,
		Users:          []models.OrgCollectionMember{},
	}
	return obj
}

func testFullyFilledFolder() models.Folder {
	obj := models.Folder{
		Name:   "sensitive-name",
		Object: models.ObjectTypeFolder,
	}
	return obj
}

func testFullyFilledItem() models.Item {
	createdDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	deletedDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	revisionDate := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	lastPasswordUsedDate := time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)

	obj := models.Item{
		Attachments: []models.Attachment{
			{
				ID:       "public-id",
				FileName: "sensitive-filename",
				Size:     "public-size",
				SizeName: "public-size-name",
				Url:      "public-url",
				Key:      "already-encrypted-key",
				Object:   models.ObjectTypeAttachment,
			},
		},
		CreationDate:  &createdDate,
		CollectionIds: []string{"test-collection-id"},
		DeletedDate:   &deletedDate,
		Edit:          true,
		Favorite:      true,
		Fields: []models.Field{
			{
				Name:  "sensitive-boolfield-name",
				Value: "sensitive-true",
				Type:  models.FieldTypeBoolean,
			},
			{
				Name:  "sensitive-hiddenfield-name",
				Value: "sensitive-hiddenfield-value",
				Type:  models.FieldTypeHidden,
			},
			{
				Name:  "sensitive-linkedfield-name",
				Value: "sensitive-linkedfield-value",
				Type:  models.FieldTypeLinked,
			},
			{
				Name:  "sensitive-textfield-name",
				Value: "sensitive-textfield-value",
				Type:  models.FieldTypeText,
			},
		},
		FolderID:            "test-folder-id",
		ID:                  "test-id",
		Login:               testFullyFilledLogin(),
		Name:                "sensitive-name",
		Notes:               "sensitive-notes",
		Object:              models.ObjectTypeItem,
		OrganizationID:      orgUuid,
		OrganizationUseTotp: true,
		PasswordHistory: []models.PasswordHistoryItem{
			{
				LastUsedDate: &lastPasswordUsedDate,
				Password:     "sensitive-password",
			},
		},
		Reprompt:     1,
		RevisionDate: &revisionDate,
		SecureNote: models.SecureNote{
			Type: 3,
		},
		Type:         models.ItemTypeLogin,
		ViewPassword: true,
	}
	return obj
}

func testFullyFilledLogin() models.Login {
	return models.Login{
		Username: "sensitive-username",
		Password: "sensitive-password",
		Totp:     "sensitive-totp",
		URIs: []models.LoginURI{
			{
				URI:   "sensitive-uri-basedomain",
				Match: models.URIMatchBaseDomain.ToPointer(),
			},
			{
				URI:   "sensitive-uri-exact",
				Match: models.URIMatchExact.ToPointer(),
			},
			{
				URI:   "sensitive-uri-never",
				Match: models.URIMatchNever.ToPointer(),
			},
			{
				URI:   "sensitive-uri-regexp",
				Match: models.URIMatchRegExp.ToPointer(),
			},
			{
				URI:   "sensitive-uri-startwith",
				Match: models.URIMatchStartWith.ToPointer(),
			},
		},
	}
}
