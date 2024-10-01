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

	vault := NewWebAPIVaultWithMockedTransport(t)

	err := vault.LoginWithPassword(ctx, "test@laverse.net", testPassword)
	if err != nil {
		t.Fatalf("vault unlock failed: %v", err)
	}

	assert.Equal(t, vault.loginAccount.VaultFormat, "API")
	assert.Equal(t, vault.loginAccount.Email, "test@laverse.net")
	assert.Equal(t, vault.loginAccount.AccountUUID, "e8dababd-242e-4900-becf-e88bc021dda8")
	assert.Equal(t, vault.loginAccount.KdfType, 0)
	assert.Equal(t, vault.loginAccount.KdfIterations, 600000)
	assert.Equal(t, vault.loginAccount.ProtectedRSAPrivateKey, "2.D2aLa8ne/DAkeSzctQISVw==|/xoGM5i5JGJTH/vohUuwTFrTx3hd/gt3kBD/FQdeLMtYjl1u96sh0ECmoERqGHeSfXj+iAb9kpTIOKG8LwmkZGUJBI90Mw0M7ODmf7E8eQ+aGF+bGqTSMQ1wtpunEyFVodlg92YN8Ddlb2V9J4uN8ykpHYNDmQiYLZ8bl6vCODRGPyzLvx5M8DbITVL5PhsjKDLLrVV8lFCgCcAL5YLfkghYhFELyX15zXA/KYEnwggDka3hG5+HHFOVZSeyk7Gi6M4TX2wADbTXz1/Wsho8oxFUrtNiOB3ZiY2cx9UWttpzMXoGfi2gJcP1db/nTfWenOLlzw6Od4VyRzsXsfyGwbqBqDnNFkjLvhjVw4JO+psF//xAMDs14101Tf2wFkB6toQ+zdnDphXUeKmiVPQ7gMnQlOWN5tWvjjmYOO4Y63sGpP24cDOdEScIdebZRSA8uOhTzadfKfOiH5zVYZzXs33FQ0li4nBrsj5xYa6PP4D1P7gqjxClPdguwkdLoZ7JvgIlyRIwEcORi5Ich8RWF/kqRBwk0QSzK1mTlHHU5xtSgi4MLNVx4qTYNaJVBtwL9d2MD9LeNn4Z2PL4A7qnszHqsERiQQDxNEgMxMBHDgSXqQbtQxRvsI6oY+yNbN7uVWw4o8AC3f5GBdxzIqcN1mgEM5ix5aDt15w3MhP2FHtf2neKI68TnL8WnT1fT8BVlbECIiUqK0tfq5aTjdSh3gCS15jvZ1H60h+K8+O/nDfquzVjY7UsTGwA+UtS8/JGiaUhc0VhxJo8P1V2VSCiu51d5q3De1vDg5R2VEgBmTchZyTIodC+3+7ACTOwkNCCdIJN7xKOcIGFA7QOuyJtBeXT4Rd9UGMHSL054IB/315WVDiwrP9W1aP0nHzFs+qAXbH5o1E+AmfMDyoHopjGgbUw22r837kHzf5Qe8QhRYPzQvDowfCPhdoy23cbN1VsNavNwTC9hcG5oYMwkK66xP3MEM9UfGD22pwxe7M8U4BRdLCCHbi95eklXE6Mg6DpWsAdMgokQbOvnlwgKfrlbltqXUE/vUQI8TB3AE1Nkt0ST4quTriMuuyiHdeeZV4UkV9jWt/5bTCfdrCYuGZP7g4shbfNcP7u1Zrdxv+EuUwGIOOTrNV5awmBnL3iHE5ya2MnqmRyfWiPIT5majZCk06yxj4XzyIPOpjYKFt0MOgLvG1GllmdtRqg7tMVvc5ZFo5KWIxLsIJD12UjA1GYYoFdX4+wsNbPjfnlE6D2PrtWUICnBFJzYpfyrKTe01k8G8hyyz+tVzBRfz8EA2ew1+hlVcAgSPCcBzhDgqPe+RSPi7ZSd66be1gDhGAftWFM8Z0MrMklXi2DyjjaKBNsZZD8qTcLcobm8nqHUQtnr5JCbmgP3rau8NY/fxeFHsvSiZQoB1aI/y+Sz/R4r+T9cg8hjmS/FUHDO+m6a6nuWNFwz8wIluM557oOTl+A9UGFF50Gpzmf97VdQjM3ZREazQ7la6AobzS3BHI6FNdxN9LTyMpYo+WODv52/VwU3ODH7wf5bz2OHZhk2NG5R7pSH7qg8jM+/MtJkFumENV0qMecozIkP6e4CyI9ua4YwI9n7G5OgKYMG1aj2PRSny2JSLS8aHF1TkRL8SD0nZFCox0=|muEtiwIuZxhuuLv0nouEdxHU2CO+I7JXKZuYHWiv/OE=")
	assert.Equal(t, vault.loginAccount.ProtectedSymmetricKey, "2.lkAJiJtCKPHFPrZ96+j2Xg==|5XJtrKUndcGy28thFukrmgMcLp+BOVdkF+KcuOnfshq9AN1PFhna9Es96CVARCnjTcWuHuqvgnGmcOHTrf8fyfLv63VBsjLgLZk8rCXJoKE=|9dwgx4/13AD+elE2vE7vlSQoe8LbCGGlui345YrKvXY=")
}

func TestObjectCreation(t *testing.T) {
	ctx := context.Background()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	vault := NewWebAPIVaultWithMockedTransport(t)

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
