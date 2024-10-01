package embedded

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/stretchr/testify/assert"
)

func TestLoginAsPasswordLoadsAccountInformation(t *testing.T) {
	ctx := context.Background()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	vault := NewMockedWebAPIVault(t, mockedClient())

	err := vault.LoginWithPassword(ctx, "test@laverse.net", testPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, "test@laverse.net", vault.loginAccount.Email)
	assert.Equal(t, "e8dababd-242e-4900-becf-e88bc021dda8", vault.loginAccount.AccountUUID)
	assert.Equal(t, models.KdfTypePBKDF2_SHA256, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 600000, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, "2.D2aLa8ne/DAkeSzctQISVw==|/xoGM5i5JGJTH/vohUuwTFrTx3hd/gt3kBD/FQdeLMtYjl1u96sh0ECmoERqGHeSfXj+iAb9kpTIOKG8LwmkZGUJBI90Mw0M7ODmf7E8eQ+aGF+bGqTSMQ1wtpunEyFVodlg92YN8Ddlb2V9J4uN8ykpHYNDmQiYLZ8bl6vCODRGPyzLvx5M8DbITVL5PhsjKDLLrVV8lFCgCcAL5YLfkghYhFELyX15zXA/KYEnwggDka3hG5+HHFOVZSeyk7Gi6M4TX2wADbTXz1/Wsho8oxFUrtNiOB3ZiY2cx9UWttpzMXoGfi2gJcP1db/nTfWenOLlzw6Od4VyRzsXsfyGwbqBqDnNFkjLvhjVw4JO+psF//xAMDs14101Tf2wFkB6toQ+zdnDphXUeKmiVPQ7gMnQlOWN5tWvjjmYOO4Y63sGpP24cDOdEScIdebZRSA8uOhTzadfKfOiH5zVYZzXs33FQ0li4nBrsj5xYa6PP4D1P7gqjxClPdguwkdLoZ7JvgIlyRIwEcORi5Ich8RWF/kqRBwk0QSzK1mTlHHU5xtSgi4MLNVx4qTYNaJVBtwL9d2MD9LeNn4Z2PL4A7qnszHqsERiQQDxNEgMxMBHDgSXqQbtQxRvsI6oY+yNbN7uVWw4o8AC3f5GBdxzIqcN1mgEM5ix5aDt15w3MhP2FHtf2neKI68TnL8WnT1fT8BVlbECIiUqK0tfq5aTjdSh3gCS15jvZ1H60h+K8+O/nDfquzVjY7UsTGwA+UtS8/JGiaUhc0VhxJo8P1V2VSCiu51d5q3De1vDg5R2VEgBmTchZyTIodC+3+7ACTOwkNCCdIJN7xKOcIGFA7QOuyJtBeXT4Rd9UGMHSL054IB/315WVDiwrP9W1aP0nHzFs+qAXbH5o1E+AmfMDyoHopjGgbUw22r837kHzf5Qe8QhRYPzQvDowfCPhdoy23cbN1VsNavNwTC9hcG5oYMwkK66xP3MEM9UfGD22pwxe7M8U4BRdLCCHbi95eklXE6Mg6DpWsAdMgokQbOvnlwgKfrlbltqXUE/vUQI8TB3AE1Nkt0ST4quTriMuuyiHdeeZV4UkV9jWt/5bTCfdrCYuGZP7g4shbfNcP7u1Zrdxv+EuUwGIOOTrNV5awmBnL3iHE5ya2MnqmRyfWiPIT5majZCk06yxj4XzyIPOpjYKFt0MOgLvG1GllmdtRqg7tMVvc5ZFo5KWIxLsIJD12UjA1GYYoFdX4+wsNbPjfnlE6D2PrtWUICnBFJzYpfyrKTe01k8G8hyyz+tVzBRfz8EA2ew1+hlVcAgSPCcBzhDgqPe+RSPi7ZSd66be1gDhGAftWFM8Z0MrMklXi2DyjjaKBNsZZD8qTcLcobm8nqHUQtnr5JCbmgP3rau8NY/fxeFHsvSiZQoB1aI/y+Sz/R4r+T9cg8hjmS/FUHDO+m6a6nuWNFwz8wIluM557oOTl+A9UGFF50Gpzmf97VdQjM3ZREazQ7la6AobzS3BHI6FNdxN9LTyMpYo+WODv52/VwU3ODH7wf5bz2OHZhk2NG5R7pSH7qg8jM+/MtJkFumENV0qMecozIkP6e4CyI9ua4YwI9n7G5OgKYMG1aj2PRSny2JSLS8aHF1TkRL8SD0nZFCox0=|muEtiwIuZxhuuLv0nouEdxHU2CO+I7JXKZuYHWiv/OE=", vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, "2.lkAJiJtCKPHFPrZ96+j2Xg==|5XJtrKUndcGy28thFukrmgMcLp+BOVdkF+KcuOnfshq9AN1PFhna9Es96CVARCnjTcWuHuqvgnGmcOHTrf8fyfLv63VBsjLgLZk8rCXJoKE=|9dwgx4/13AD+elE2vE7vlSQoe8LbCGGlui345YrKvXY=", vault.loginAccount.ProtectedSymmetricKey)
}

func TestLoginAsPasswordLoadsAccountInformationForArgon2(t *testing.T) {
	ctx := context.Background()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	vault := NewMockedWebAPIVault(t, mockedClientArgon2())

	err := vault.LoginWithPassword(ctx, "test-202341-argon2@laverse.net", "test1234")
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, "API", vault.loginAccount.VaultFormat)
	assert.Equal(t, "test-202341-argon2@laverse.net", vault.loginAccount.Email)
	assert.Equal(t, "c64a1f8d-223c-40fd-adde-a90d58dbc26c", vault.loginAccount.AccountUUID)
	assert.Equal(t, models.KdfTypeArgon2, vault.loginAccount.KdfConfig.KdfType)
	assert.Equal(t, 3, vault.loginAccount.KdfConfig.KdfIterations)
	assert.Equal(t, "2.0Xc+uV+6YVa+Zkg74C/MJA==|xTbZGy2r8/kHcXXy23YDHWDTHjasUUoHiWJ84rGEhuPD27GTF8DrFL+lbERo3+OH7MXvseAFygKArvSumlCQnMhT/dswJMH8ZEMuxQUZ8I7kg/7eNqxRndxyC8alZh3VM4FC7TsuVBe6BF4pbPQYX4etFL8yOHLpIhwkKcG7+IfQHQHwOXpsSiGpADQlk5er990snhAXwGpgROqoveO9klpNJXuEzJKMFMdoo9FCaRsJ4bB3BE/Y+8Ph7ek8mGyoUUqNFp8Z/T+XN+9kvxDFnVFtYp1p+sU8LpCSvaq3dAImQt6X7vAgfjbWVkTxB0HlBdMkjpg8BAK7qpscT5oH83SZqdNPRi9kkT9Y30xl0QvvJzXLcjqOS6je+i3vB1r4O8X3ZIj+th4cfZOfHKpzDLbGyAezxiiWwQ4xt0USQ9rLv+BU5YNDKnibmsIooKHuh8wV0qht8vr2CEVavXOmmBi/6bGIWJMOs2K52be5LJkYL+653dsKXBN4uahQDMdfs9vPUA7OoIFR9BZvQiGMDstFYVgJUulOkj2J4nieuTAmurQe6nS6U4v1swI7DIzdo8ZCJcAQfWYWk8IBwh6gQxSuPMg7+O2dntbPFkMP2KhoisUucoXNIDXmxwYAMeP01KMNSgFc6hL6WnLzyVcRlSLy7OC9qwTD6ZH6XA4D/MO9MINTTzw4+/pZvwTeuXMqB0HkcjTpWzTUi405uV6qh8CxYOaNEl6kNAMiuVNASToqYb/EBC0uGOybOzcNZOCiXmuiPVOstiz65KP1d3Dfl9al48hpxnDxFyFub5NDx8xADA3SzOlI9xMCsC/mcY0/fRlwoTDFCLzfdtPIdSn5N8/YySY10e/TXftKGV7bLvzGAlOntwoiWaLJbyHAnEXNUVFxJu5XVaVccXcxIKete42S291dCu4yk35KsIQU0jBaPB/hGXLJvvySl9/kARl0zIXJH+Pk0hlk+/IRH9HNZrru93WTgW9KjEeT2vaZ6UqULkkspIfoUrFxQfSyAxycDaqa4EHt1QJBKyC9+aWEOXNwtxkJTnhtvlqRPOGUpdkAyC5ebfdX2URZnd/2TR/PTviWaDKe/g51UWr8QepgwbJjkBeuZMSPCNHkcLdmwEZERisbZ7H50hDQhhK+qfmMOvkVrRRXKT2ICbQzchm5rDasdtMed7vIf3beS0ESIR8kZ1WmteI9dvWmagnXRlfsUZW0y0KWe9Ma/ISOa6QPb0Cc7T8LUg2lpREQpRTt119RX9Xugd2uZ+UzbGpqn9u5JY2+vddwv+zFL0g/Vu9h3kd02v+WTXeV8wLDJYkoc6XeTBzC0uQtNJt4oN7hWcMBrkYmnXxslIjGw/nE7eBnG02XWTCeTjADqXtz2j4xMFmzJod+4j4f6MQZBII9Sz5A2FfSbadUuukEYcFvkqLnjG/sMzcBQMIDP6vmy7VRnETahlNi7Pt+gRLHuhGN1BbScolW7a0YnWlI1T0MtKb4DBsjX9GfotxkG+ORg/i45YnG6KPQ6+jgXU3Xtf0tgjTqHVgc7TOCm9R2b4G24RSQXx3WaMP1slbSYsf/YmXIqdg/dVaCFrJdA1v2bDFymuPbA0P4Ey8c4ylInXGgK4yxS7nVC4qbbVtK5bWT3SRaBWw=|0r9TcWfl93tZln8+ZGrPwZTHBbrLiSuVXvoI+cQeSr0=", vault.loginAccount.ProtectedRSAPrivateKey)
	assert.Equal(t, "2.VBpAJrIHYLv60UTYy5e7sA==|WWsozKnKPVvBYttXtlAzpmZPoffwlY3+8Wup9SBdGcO8T4Ybj808DubDhICTPMz9RliDSJ1OuaivBC0rh/7EcWV1s9KuRth5J3XFWJqmtXU=|l1ly+Uek1j4k2xQNT8iJ++GdwMTWRaSBP5mFvgJ+CGU=", vault.loginAccount.ProtectedSymmetricKey)
}

func TestObjectCreation(t *testing.T) {
	ctx := context.Background()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	vault := NewMockedWebAPIVault(t, mockedClient())

	err := vault.LoginWithPassword(ctx, "test@laverse.net", testPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	obj, err := vault.CreateObject(ctx, models.Object{
		Object: models.ObjectTypeItem,
		Type:   models.ItemTypeLogin,
		Name:   "test",
	})
	assert.NoError(t, err)
	if !assert.NotNil(t, obj) {
		return
	}

	assert.Equal(t, obj.Name, "Item in own Vault")
	assert.Equal(t, obj.Login.Username, "my-username")
}
