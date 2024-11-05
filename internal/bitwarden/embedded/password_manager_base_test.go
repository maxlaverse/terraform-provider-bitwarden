package embedded

import (
	"context"
	"crypto/x509"
	"encoding/pem"
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

var (
	testAccountPbkdf2 = Account{
		AccountUUID: "e8dababd-242e-4900-becf-e88bc021dda8",
		Email:       Pdkdf2Email,
		VaultFormat: "API",
		KdfConfig: models.KdfConfiguration{
			KdfType:       models.KdfTypePBKDF2_SHA256,
			KdfIterations: 600000,
		},
		ProtectedSymmetricKey:  Pdkdf2ProtectedSymmetricKey,
		ProtectedRSAPrivateKey: Pdkdf2ProtectedRSAPrivateKey,
	}

	testAccountArgon2 = Account{
		AccountUUID: "e8dababd-242e-4900-becf-e88bc021dda8",
		Email:       Argon2Email,
		VaultFormat: "API",
		KdfConfig: models.KdfConfiguration{
			KdfType:        models.KdfTypeArgon2,
			KdfIterations:  3,
			KdfMemory:      64,
			KdfParallelism: 4,
		},
		ProtectedSymmetricKey:  Argon2ProtectedSymmetricKey,
		ProtectedRSAPrivateKey: Argon2ProtectedRSAPrivateKey,
	}
)

func TestDecryptAccountSecretPbkdf2(t *testing.T) {
	accountSecrets, err := decryptAccountSecrets(testAccountPbkdf2, TestPassword)
	assert.NoError(t, err)
	assert.Equal(t, "jHZYmFOOr0KGorSsmWeuGMHWJDqRrz7uwZNBJkZaupM=", accountSecrets.MasterPasswordHash)

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
	accountSecrets, err := decryptAccountSecrets(testAccountArgon2, TestPassword)
	assert.NoError(t, err)
	assert.Equal(t, "3fBImY0XFvRrUSP/fe6mqUc1bjhWBuvHYJvlwnxS0i4=", accountSecrets.MasterPasswordHash)

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
	_, err := decryptAccountSecrets(testAccountPbkdf2, "wrong-password")
	assert.Error(t, err, "decryption should fail with wrong password")
	assert.ErrorIs(t, err, models.ErrWrongMasterPassword)
}

func TestEncryptItem(t *testing.T) {
	accountSecrets := computeTestAccountSecrets(t)

	objectToEncrypt := testFullyFilledObject()
	objectToEncrypt.OrganizationID = "81cc1652-dc80-472d-909f-9539d057068b"
	newObj, err := encryptItem(context.Background(), objectToEncrypt, *accountSecrets, true)
	assert.Nil(t, err)

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

	if assert.Len(t, newObj.Fields, 4) {
		assert.Equal(t, "test-boolfield-name", objectToEncrypt.Fields[0].Name)
		assertEncryptedValueOf(t, "test-boolfield-name", newObj.Fields[0].Name, *r)
		assert.Equal(t, "true", objectToEncrypt.Fields[0].Value)
		assertEncryptedValueOf(t, "true", newObj.Fields[0].Value, *r)
		assert.Equal(t, models.FieldTypeBoolean, objectToEncrypt.Fields[0].Type)
		assert.Equal(t, models.FieldTypeBoolean, newObj.Fields[0].Type)

		assert.Equal(t, "test-hiddenfield-name", objectToEncrypt.Fields[1].Name)
		assertEncryptedValueOf(t, "test-hiddenfield-name", newObj.Fields[1].Name, *r)
		assert.Equal(t, "test-hiddenfield-value", objectToEncrypt.Fields[1].Value)
		assertEncryptedValueOf(t, "test-hiddenfield-value", newObj.Fields[1].Value, *r)
		assert.Equal(t, models.FieldTypeHidden, objectToEncrypt.Fields[1].Type)
		assert.Equal(t, models.FieldTypeHidden, newObj.Fields[1].Type)

		assert.Equal(t, "test-linkedfield-name", objectToEncrypt.Fields[2].Name)
		assertEncryptedValueOf(t, "test-linkedfield-name", newObj.Fields[2].Name, *r)
		assert.Equal(t, "test-linkedfield-value", objectToEncrypt.Fields[2].Value)
		assertEncryptedValueOf(t, "test-linkedfield-value", newObj.Fields[2].Value, *r)
		assert.Equal(t, models.FieldTypeLinked, objectToEncrypt.Fields[2].Type)
		assert.Equal(t, models.FieldTypeLinked, newObj.Fields[2].Type)

		assert.Equal(t, "test-textfield-name", objectToEncrypt.Fields[3].Name)
		assertEncryptedValueOf(t, "test-textfield-name", newObj.Fields[3].Name, *r)
		assert.Equal(t, "test-textfield-value", objectToEncrypt.Fields[3].Value)
		assertEncryptedValueOf(t, "test-textfield-value", newObj.Fields[3].Value, *r)
		assert.Equal(t, models.FieldTypeText, objectToEncrypt.Fields[3].Type)
		assert.Equal(t, models.FieldTypeText, newObj.Fields[3].Type)
	}

	assert.Equal(t, "test-folder-id", objectToEncrypt.FolderID)
	assert.Equal(t, "test-folder-id", newObj.FolderID)

	assert.Equal(t, "test-id", objectToEncrypt.ID)
	assert.Equal(t, "test-id", newObj.ID)

	assert.Equal(t, "test-username", objectToEncrypt.Login.Username)
	assertEncryptedValueOf(t, "test-username", newObj.Login.Username, *r)

	assert.Equal(t, "test-password", objectToEncrypt.Login.Password)
	assertEncryptedValueOf(t, "test-password", newObj.Login.Password, *r)

	assert.Equal(t, "test-totp", objectToEncrypt.Login.Totp)
	assertEncryptedValueOf(t, "test-totp", newObj.Login.Totp, *r)

	if assert.Len(t, newObj.Login.URIs, 5) {
		assert.Equal(t, "test-uri-basedomain", objectToEncrypt.Login.URIs[0].URI)
		assertEncryptedValueOf(t, "test-uri-basedomain", newObj.Login.URIs[0].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[0].Match) {
			assert.Equal(t, models.URIMatchBaseDomain, *objectToEncrypt.Login.URIs[0].Match)
			assert.Equal(t, models.URIMatchBaseDomain, *newObj.Login.URIs[0].Match)
		}

		assert.Equal(t, "test-uri-exact", objectToEncrypt.Login.URIs[1].URI)
		assertEncryptedValueOf(t, "test-uri-exact", newObj.Login.URIs[1].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[1].Match) {
			assert.Equal(t, models.URIMatchExact, *objectToEncrypt.Login.URIs[1].Match)
			assert.Equal(t, models.URIMatchExact, *newObj.Login.URIs[1].Match)
		}

		assert.Equal(t, "test-uri-never", objectToEncrypt.Login.URIs[2].URI)
		assertEncryptedValueOf(t, "test-uri-never", newObj.Login.URIs[2].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[2].Match) {
			assert.Equal(t, models.URIMatchNever, *objectToEncrypt.Login.URIs[2].Match)
			assert.Equal(t, models.URIMatchNever, *newObj.Login.URIs[2].Match)
		}

		assert.Equal(t, "test-uri-regexp", objectToEncrypt.Login.URIs[3].URI)
		assertEncryptedValueOf(t, "test-uri-regexp", newObj.Login.URIs[3].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[3].Match) {
			assert.Equal(t, models.URIMatchRegExp, *objectToEncrypt.Login.URIs[3].Match)
			assert.Equal(t, models.URIMatchRegExp, *newObj.Login.URIs[3].Match)
		}

		assert.Equal(t, "test-uri-startwith", objectToEncrypt.Login.URIs[4].URI)
		assertEncryptedValueOf(t, "test-uri-startwith", newObj.Login.URIs[4].URI, *r)
		if assert.NotNil(t, objectToEncrypt.Login.URIs[4].Match) {
			assert.Equal(t, models.URIMatchStartWith, *objectToEncrypt.Login.URIs[4].Match)
			assert.Equal(t, models.URIMatchStartWith, *newObj.Login.URIs[4].Match)
		}
	}

	assert.Equal(t, "test-name", objectToEncrypt.Name)
	assertEncryptedValueOf(t, "test-name", newObj.Name, *r)

	assert.Equal(t, "test-notes", objectToEncrypt.Notes)
	assertEncryptedValueOf(t, "test-notes", newObj.Notes, *r)

	assert.Equal(t, models.ObjectTypeItem, objectToEncrypt.Object)
	assert.Equal(t, models.ObjectTypeItem, newObj.Object)

	assert.Equal(t, "81cc1652-dc80-472d-909f-9539d057068b", objectToEncrypt.OrganizationID)
	assert.Equal(t, "81cc1652-dc80-472d-909f-9539d057068b", newObj.OrganizationID)

	assert.Equal(t, true, objectToEncrypt.OrganizationUseTotp)
	assert.Equal(t, true, newObj.OrganizationUseTotp)

	if assert.Len(t, newObj.PasswordHistory, 1) {
		assert.Equal(t, "test-password", objectToEncrypt.PasswordHistory[0].Password)
		assertEncryptedValueOf(t, "test-password", newObj.PasswordHistory[0].Password, *r)

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

}

func assertEncryptedValueOf(t *testing.T, expected, value string, k symmetrickey.Key) {
	out, err := decryptStringAsBytes(value, k)
	assert.Nil(t, err)
	assert.Equal(t, expected, string(out))
}

func testFullyFilledObject() models.Object {
	createdDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	deletedDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	revisionDate := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	lastPasswordUsedDate := time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)

	obj := models.Object{
		CreationDate:  &createdDate,
		CollectionIds: []string{"test-collection-id"},
		DeletedDate:   &deletedDate,
		Edit:          true,
		Favorite:      true,
		Fields: []models.Field{
			{
				Name:  "test-boolfield-name",
				Value: "true",
				Type:  models.FieldTypeBoolean,
			},
			{
				Name:  "test-hiddenfield-name",
				Value: "test-hiddenfield-value",
				Type:  models.FieldTypeHidden,
			},
			{
				Name:  "test-linkedfield-name",
				Value: "test-linkedfield-value",
				Type:  models.FieldTypeLinked,
			},
			{
				Name:  "test-textfield-name",
				Value: "test-textfield-value",
				Type:  models.FieldTypeText,
			},
		},
		FolderID:            "test-folder-id",
		ID:                  "test-id",
		Login:               testFullyFilledLogin(),
		Name:                "test-name",
		Notes:               "test-notes",
		Object:              models.ObjectTypeItem,
		OrganizationID:      "81cc1652-dc80-472d-909f-9539d057068b",
		OrganizationUseTotp: true,
		PasswordHistory: []models.PasswordHistoryItem{
			{
				LastUsedDate: &lastPasswordUsedDate,
				Password:     "test-password",
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
		Username: "test-username",
		Password: "test-password",
		Totp:     "test-totp",
		URIs: []models.LoginURI{
			{
				URI:   "test-uri-basedomain",
				Match: models.URIMatchBaseDomain.ToPointer(),
			},
			{
				URI:   "test-uri-exact",
				Match: models.URIMatchExact.ToPointer(),
			},
			{
				URI:   "test-uri-never",
				Match: models.URIMatchNever.ToPointer(),
			},
			{
				URI:   "test-uri-regexp",
				Match: models.URIMatchRegExp.ToPointer(),
			},
			{
				URI:   "test-uri-startwith",
				Match: models.URIMatchStartWith.ToPointer(),
			},
		},
	}
}

func computeTestAccountSecrets(t *testing.T) *AccountSecrets {
	accountSecrets, err := decryptAccountSecrets(testAccountPbkdf2, TestPassword)
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
