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

var (
	testAccount = Account{
		AccountUUID: "e8dababd-242e-4900-becf-e88bc021dda8",
		Email:       "test@laverse.net",
		VaultFormat: "API",
		KdfConfig: models.KdfConfiguration{
			KdfType:       models.KdfTypePBKDF2_SHA256,
			KdfIterations: 600000,
		},
		ProtectedSymmetricKey:  "2.lkAJiJtCKPHFPrZ96+j2Xg==|5XJtrKUndcGy28thFukrmgMcLp+BOVdkF+KcuOnfshq9AN1PFhna9Es96CVARCnjTcWuHuqvgnGmcOHTrf8fyfLv63VBsjLgLZk8rCXJoKE=|9dwgx4/13AD+elE2vE7vlSQoe8LbCGGlui345YrKvXY=",
		ProtectedRSAPrivateKey: "2.D2aLa8ne/DAkeSzctQISVw==|/xoGM5i5JGJTH/vohUuwTFrTx3hd/gt3kBD/FQdeLMtYjl1u96sh0ECmoERqGHeSfXj+iAb9kpTIOKG8LwmkZGUJBI90Mw0M7ODmf7E8eQ+aGF+bGqTSMQ1wtpunEyFVodlg92YN8Ddlb2V9J4uN8ykpHYNDmQiYLZ8bl6vCODRGPyzLvx5M8DbITVL5PhsjKDLLrVV8lFCgCcAL5YLfkghYhFELyX15zXA/KYEnwggDka3hG5+HHFOVZSeyk7Gi6M4TX2wADbTXz1/Wsho8oxFUrtNiOB3ZiY2cx9UWttpzMXoGfi2gJcP1db/nTfWenOLlzw6Od4VyRzsXsfyGwbqBqDnNFkjLvhjVw4JO+psF//xAMDs14101Tf2wFkB6toQ+zdnDphXUeKmiVPQ7gMnQlOWN5tWvjjmYOO4Y63sGpP24cDOdEScIdebZRSA8uOhTzadfKfOiH5zVYZzXs33FQ0li4nBrsj5xYa6PP4D1P7gqjxClPdguwkdLoZ7JvgIlyRIwEcORi5Ich8RWF/kqRBwk0QSzK1mTlHHU5xtSgi4MLNVx4qTYNaJVBtwL9d2MD9LeNn4Z2PL4A7qnszHqsERiQQDxNEgMxMBHDgSXqQbtQxRvsI6oY+yNbN7uVWw4o8AC3f5GBdxzIqcN1mgEM5ix5aDt15w3MhP2FHtf2neKI68TnL8WnT1fT8BVlbECIiUqK0tfq5aTjdSh3gCS15jvZ1H60h+K8+O/nDfquzVjY7UsTGwA+UtS8/JGiaUhc0VhxJo8P1V2VSCiu51d5q3De1vDg5R2VEgBmTchZyTIodC+3+7ACTOwkNCCdIJN7xKOcIGFA7QOuyJtBeXT4Rd9UGMHSL054IB/315WVDiwrP9W1aP0nHzFs+qAXbH5o1E+AmfMDyoHopjGgbUw22r837kHzf5Qe8QhRYPzQvDowfCPhdoy23cbN1VsNavNwTC9hcG5oYMwkK66xP3MEM9UfGD22pwxe7M8U4BRdLCCHbi95eklXE6Mg6DpWsAdMgokQbOvnlwgKfrlbltqXUE/vUQI8TB3AE1Nkt0ST4quTriMuuyiHdeeZV4UkV9jWt/5bTCfdrCYuGZP7g4shbfNcP7u1Zrdxv+EuUwGIOOTrNV5awmBnL3iHE5ya2MnqmRyfWiPIT5majZCk06yxj4XzyIPOpjYKFt0MOgLvG1GllmdtRqg7tMVvc5ZFo5KWIxLsIJD12UjA1GYYoFdX4+wsNbPjfnlE6D2PrtWUICnBFJzYpfyrKTe01k8G8hyyz+tVzBRfz8EA2ew1+hlVcAgSPCcBzhDgqPe+RSPi7ZSd66be1gDhGAftWFM8Z0MrMklXi2DyjjaKBNsZZD8qTcLcobm8nqHUQtnr5JCbmgP3rau8NY/fxeFHsvSiZQoB1aI/y+Sz/R4r+T9cg8hjmS/FUHDO+m6a6nuWNFwz8wIluM557oOTl+A9UGFF50Gpzmf97VdQjM3ZREazQ7la6AobzS3BHI6FNdxN9LTyMpYo+WODv52/VwU3ODH7wf5bz2OHZhk2NG5R7pSH7qg8jM+/MtJkFumENV0qMecozIkP6e4CyI9ua4YwI9n7G5OgKYMG1aj2PRSny2JSLS8aHF1TkRL8SD0nZFCox0=|muEtiwIuZxhuuLv0nouEdxHU2CO+I7JXKZuYHWiv/OE=",
	}

	rsaPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAlRbtt5Vyku3dUIkPR0Y3v94qvZReuwAijIwHQGQOuKw6lKVV
HL29rZ93TwCG2P5a+GKH2+2fIbT/wTMTK4K1ElxQZ/2yLN9Hfu2d/ITNTsfTzPXv
F3fao3Q0JmD7DNJS2bJqng3so28aFddOZ03H55m9T6+0ZJqrMdgE5Z1V/4I7LFF6
JxGGZ4mg99OvfQ5K8GOBM6SCI6h5eMXM5EkSM9vol9sRxvVLZmvNKH3UP9riyQwt
dcD5IxmY1y34Bg4b+a8tYaP7v90xF73uKs+287yNPLhWE9i+/gXwvApVxENG9SCP
lSrAEHd1NZPfsZhHoG+LXhyCZu2COttZess44QIDAQABAoIBAAIXgN54+qMpJ+2M
yGsdvGj3vCy9+vSyWi8Tr3icdXMrKfTVgMUlvEurcOI/Mcp+v55MF3JF0kwylh3N
pSwbV3DBHN5Hp5xu8HmtahsoWRnXo98Z4oOB7U5gAj5kBmkMtKhB/fJW+UzF/C/b
I2906Tw59Uy2XIsROzvjMeGPnddh1LbvXUb9nAmhi7napdwCUbeqvatu56GyiXxV
03DTwhbfuU+nMi/M556WPEkPbJoG4bF82WqQ+6+a1NfE2wg//cc5CzXQehC6jGx6
Pi+uNUtPhMSyTnJgpvg+Ob2r/LTiL0zic00ka29xUsi7EwXKUR2ih8JTkSPaTR06
3ezrg3ECgYEAzi/MnXfWp49jRgb4bHE9Cy63hehgaBvIcEyhFKOz5OD6nI4Lg/3z
SNDQo9YhwMqschqQLHVEtjxDT0V/RHdX4icTF+zSCl/T79EtM/R1nMT0MSIXV7IE
NtPbnqXOjrbe3vgjLvBst/cWpGHiML+znCqukHOevSn7yUlg4b1aMVECgYEAuRvL
YnbNlps/nql1EW9DUKOEV7kBvehwGzYpFfZ7pRscl6RyISTipGMOzJmOSfscYwqR
HrFpTNMNxjnyXOuv4OTC7bCIUJc6N8AZ4jm4452ibxpzktlO8Im+TbsD6mZDT8zB
d+8o/8ST9j1zy43Sb+f5vxfB6fC9vUXpBW0+KpECgYEAzQKT9cJxSVv1/mvx2Ilj
g9nompmqOfnd+2MGCuqWdS4JoV5PLudzXeRaf3zrRLGAc1fcIIhdUMFsv8Y/O8la
NcBaaMCNO8l6hoo64tzfkIf4sV3PTd/v9sACL6V3U0mbIqIhAYwG3YguGDZHW+dQ
ZCfAOFrt6/Jxqvtt/CZ1JnECgYBOxmdNZeWj/Dmc2dy6KLFq9ctyUYdOPEbJLcla
UWTZJKqMVi1DsaDJ+GXp6EdHcJfqBisv9qwrR34LJ8nehWZ5vKC/6mp4cYMTCqt5
PLtUEld4FLeufNA9SUE1bysBa7ellCuZUKwP/KZDGm/W5mnxubTs/71EQ3FbxQ6f
gpf8IQKBgGHK8j8rvxtszoQKUY+XpWFrP0x5pDLiAkmQ0bmF2KRIahq3anla6n3i
/LL5BrUdMjEnvAb+RASq+41rceq4rLcz0pA2yOWNjhbCAPFdU5MQMkJ4/zqHtzvd
GqwE00g9gizQ6CmsaNNJh7y6gNg0TBU2EGqTaQMz37fheAEt3NSt
-----END RSA PRIVATE KEY-----
`

	encryptionKey = "Vr+KA/il3QX4z7EqFnhQ3U8TtETlQPKkXHCE2PiR75wwzDVRutR4rib/jMtgZ1S/gPyOEXbwKFju2oJq3njVLg=="
	orgKey        = "4.JW3mktbL7vpTVRweZdQBAirJuAEhSRn37zcXZjDjI47weFKkeZkvPxZWCqFYC/P5qCJwEYMbv7lTETkWDg6paevVfhJ35buGcTQdEbQxAJebzPahEcUstj11l4Y9T5RaDiAJR8+drrGJ3fKV3v3hymKz2o9fUfK1epuLFll2nnWSOjCcuRe/+zz5VwIVx4WJAPJHmiS6eofbj/DTIQCzG4JkR0UzT66ouLcgmPL1nGOqVI7KxRpL5yVj75UkjniHkWAcB7lfAxWXw2GhDJ/2L685uA3820ItTbxjCwLQOvjBttgrbURmkeP9BD+KkO4V6vb8bbTWNSvggXKk2h1CMw=="
	testPassword  = "test12341234"
)

func TestDecryptAccountSecret(t *testing.T) {
	accountSecrets, err := decryptAccountSecrets(testAccount, testPassword)
	assert.NoError(t, err)
	assert.Equal(t, "qOOJSiS6KGqePb+ZxBPD9G37cZjFfViArWiHCd0koK4=", accountSecrets.MasterPasswordHash)

	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(accountSecrets.RSAPrivateKey),
		},
	)

	assert.Equal(t, rsaPrivateKey, strings.Replace(string(pemdata), "\\n", "\n", -1))
	assert.Contains(t, accountSecrets.MainKey.Summary(), encryptionKey)
}

func TestDecryptAccountSecretWrongPassword(t *testing.T) {
	_, err := decryptAccountSecrets(testAccount, "wrong-password")
	assert.Error(t, err, "decryption should fail with wrong password")
	assert.ErrorIs(t, err, models.ErrWrongMasterPassword)
}

func TestEncryptItem(t *testing.T) {
	accountSecrets := testAccountSecrets(t)

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

func testAccountSecrets(t *testing.T) *AccountSecrets {
	accountSecrets, err := decryptAccountSecrets(testAccount, testPassword)
	if err != nil {
		t.Fatal(err)
	}

	key, err := decryptOrganizationKey(orgKey, accountSecrets.RSAPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	accountSecrets.OrganizationSecrets = map[string]OrganizationSecret{
		"81cc1652-dc80-472d-909f-9539d057068b": {
			OrganizationUUID: "81cc1652-dc80-472d-909f-9539d057068b",
			Key:              *key,
		},
	}
	return accountSecrets
}
