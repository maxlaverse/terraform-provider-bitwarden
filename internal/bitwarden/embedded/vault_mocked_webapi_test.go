package embedded

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/keybuilder"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

func NewMockedWebAPIVault(t *testing.T, client webapi.Client) webAPIVault {
	vault := NewWebAPIVault("http://127.0.0.1:8081/").(*webAPIVault)
	vault.client = client
	return *vault
}

func mockedClient() webapi.Client {
	client := http.Client{Transport: httpmock.DefaultTransport}
	httpmock.RegisterResponder("POST", "http://127.0.0.1:8081/identity/accounts/prelogin",
		httpmock.NewStringResponder(200, `{"kdf":0,"kdfIterations":600000,"kdfMemory":null,"kdfParallelism":null}`))

	// Regexp match (could use httpmock.RegisterRegexpResponder instead)
	httpmock.RegisterResponder("POST", `http://127.0.0.1:8081/identity/connect/token`,
		httpmock.NewStringResponder(200, `{"ForcePasswordReset":false,"Kdf":0,"KdfIterations":600000,"KdfMemory":null,"KdfParallelism":null,"Key":"2.lkAJiJtCKPHFPrZ96+j2Xg==|5XJtrKUndcGy28thFukrmgMcLp+BOVdkF+KcuOnfshq9AN1PFhna9Es96CVARCnjTcWuHuqvgnGmcOHTrf8fyfLv63VBsjLgLZk8rCXJoKE=|9dwgx4/13AD+elE2vE7vlSQoe8LbCGGlui345YrKvXY=","MasterPasswordPolicy":{"object":"masterPasswordPolicy"},"PrivateKey":"2.D2aLa8ne/DAkeSzctQISVw==|/xoGM5i5JGJTH/vohUuwTFrTx3hd/gt3kBD/FQdeLMtYjl1u96sh0ECmoERqGHeSfXj+iAb9kpTIOKG8LwmkZGUJBI90Mw0M7ODmf7E8eQ+aGF+bGqTSMQ1wtpunEyFVodlg92YN8Ddlb2V9J4uN8ykpHYNDmQiYLZ8bl6vCODRGPyzLvx5M8DbITVL5PhsjKDLLrVV8lFCgCcAL5YLfkghYhFELyX15zXA/KYEnwggDka3hG5+HHFOVZSeyk7Gi6M4TX2wADbTXz1/Wsho8oxFUrtNiOB3ZiY2cx9UWttpzMXoGfi2gJcP1db/nTfWenOLlzw6Od4VyRzsXsfyGwbqBqDnNFkjLvhjVw4JO+psF//xAMDs14101Tf2wFkB6toQ+zdnDphXUeKmiVPQ7gMnQlOWN5tWvjjmYOO4Y63sGpP24cDOdEScIdebZRSA8uOhTzadfKfOiH5zVYZzXs33FQ0li4nBrsj5xYa6PP4D1P7gqjxClPdguwkdLoZ7JvgIlyRIwEcORi5Ich8RWF/kqRBwk0QSzK1mTlHHU5xtSgi4MLNVx4qTYNaJVBtwL9d2MD9LeNn4Z2PL4A7qnszHqsERiQQDxNEgMxMBHDgSXqQbtQxRvsI6oY+yNbN7uVWw4o8AC3f5GBdxzIqcN1mgEM5ix5aDt15w3MhP2FHtf2neKI68TnL8WnT1fT8BVlbECIiUqK0tfq5aTjdSh3gCS15jvZ1H60h+K8+O/nDfquzVjY7UsTGwA+UtS8/JGiaUhc0VhxJo8P1V2VSCiu51d5q3De1vDg5R2VEgBmTchZyTIodC+3+7ACTOwkNCCdIJN7xKOcIGFA7QOuyJtBeXT4Rd9UGMHSL054IB/315WVDiwrP9W1aP0nHzFs+qAXbH5o1E+AmfMDyoHopjGgbUw22r837kHzf5Qe8QhRYPzQvDowfCPhdoy23cbN1VsNavNwTC9hcG5oYMwkK66xP3MEM9UfGD22pwxe7M8U4BRdLCCHbi95eklXE6Mg6DpWsAdMgokQbOvnlwgKfrlbltqXUE/vUQI8TB3AE1Nkt0ST4quTriMuuyiHdeeZV4UkV9jWt/5bTCfdrCYuGZP7g4shbfNcP7u1Zrdxv+EuUwGIOOTrNV5awmBnL3iHE5ya2MnqmRyfWiPIT5majZCk06yxj4XzyIPOpjYKFt0MOgLvG1GllmdtRqg7tMVvc5ZFo5KWIxLsIJD12UjA1GYYoFdX4+wsNbPjfnlE6D2PrtWUICnBFJzYpfyrKTe01k8G8hyyz+tVzBRfz8EA2ew1+hlVcAgSPCcBzhDgqPe+RSPi7ZSd66be1gDhGAftWFM8Z0MrMklXi2DyjjaKBNsZZD8qTcLcobm8nqHUQtnr5JCbmgP3rau8NY/fxeFHsvSiZQoB1aI/y+Sz/R4r+T9cg8hjmS/FUHDO+m6a6nuWNFwz8wIluM557oOTl+A9UGFF50Gpzmf97VdQjM3ZREazQ7la6AobzS3BHI6FNdxN9LTyMpYo+WODv52/VwU3ODH7wf5bz2OHZhk2NG5R7pSH7qg8jM+/MtJkFumENV0qMecozIkP6e4CyI9ua4YwI9n7G5OgKYMG1aj2PRSny2JSLS8aHF1TkRL8SD0nZFCox0=|muEtiwIuZxhuuLv0nouEdxHU2CO+I7JXKZuYHWiv/OE=","ResetMasterPassword":false,"UserDecryptionOptions":{"HasMasterPassword":true,"Object":"userDecryptionOptions"},"access_token":"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJuYmYiOjE3Mjc1MTI0NjIsImV4cCI6MTcyNzUxOTY2MiwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdHxsb2dpbiIsInN1YiI6ImU4ZGFiYWJkLTI0MmUtNDkwMC1iZWNmLWU4OGJjMDIxZGRhOCIsInByZW1pdW0iOnRydWUsIm5hbWUiOiJUZXN0IExhdmVyc2UiLCJlbWFpbCI6InRlc3RAbGF2ZXJzZS5uZXQiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwic3N0YW1wIjoiNDAzM2NmMzUtYmM1ZC00ZWNkLTk1MWMtZGJlNTQ5ZjI3NzQwIiwiZGV2aWNlIjoiNWQ5MGI0NzAtNWQxZC00NTJkLTkzNWMtYjczMGMxNzdhOGQ2Iiwic2NvcGUiOlsiYXBpIiwib2ZmbGluZV9hY2Nlc3MiXSwiYW1yIjpbIkFwcGxpY2F0aW9uIl19.hKXO3p4am134f74wx5WqmuRCT__1CcqiZ5YkU5q8FvXNY-oe_pOrU7gxpf9TDBIOQVTjVkE09dgzEKco_qMnRTT7cvQ95o3k82gNDNgyi0Yjqtv5eS8K7m7MrSOUqe6LwZwpKy3G8vDnu1O2kiN-So44cCBKQzqOrR4oimmYIMt8IWqU2qa8ZaFe1E70nv5NjUxiE_NkT2FF6M8MA7z_aEkcBeleknlKRL-gaqYqAPWhKF_msW9e3SKvm_83wt1JRwr60T4aKXbl7rB4UrSqs_lqhdb3i1Ul5_TYQXbg4yfETgT8Uh37kXtnK_8MlxNBpA8pEAnp_29Odync3T7naQ","expires_in":7200,"refresh_token":"6wCk9N0MG-abyrmBUn5m1ACq-oRnhj6mbjCeuOz_HPyPERfK8DPeTJVf3c6KtP-bIbL6mErO-ggMs_MNS2GlIg==","scope":"api offline_access","token_type":"Bearer","unofficialServer":true}`))

	httpmock.RegisterResponder("GET", `http://127.0.0.1:8081/api/sync?excludeDomains=true`,
		httpmock.NewStringResponder(200, `{
  "ciphers": [
    {
      "attachments": null,
      "card": null,
      "collectionIds": [],
      "creationDate": "2024-09-22T11:13:40.346903Z",
      "data": {
        "autofillOnPageLoad": null,
        "fields": [
          {
            "linkedId": null,
            "name": "2.svtF3aK9R1MLHt2GwH7Ykw==|PqVDa02T0Vzx6ogTNz1Xyg==|g9/Gd5OOLkVtyVAk9vkRCY2reWTZjdY1D3XsVpBfQA4=",
            "type": 0,
            "value": "2.K49DvcymvaQ/+V/3soCb1Q==|r+b2dz0YqFeA8BrHii5a0DIjkwxdwlY+aEEH2d2NdC0=|KEJAjqImC0Jit9oaFN2rSZHyetxuD7jEjMt07HPFLvU="
          }
        ],
        "name": "2.LjR8NxRtCB1noDILxNKmPQ==|s4e2AO4I3HqmPRRsbsYl0XYSuWu7+sn2K3+jNiFjsW0=|3NQW64/3RmPxe3hhUGeMTrSy9Ruh5hYRlJxIhhWhhSI=",
        "notes": "2.ke3IFCPe50UCM2XdJDS/VQ==|ksO+dyEuRBgrpAelfhxNhw==|JrbMU58V5QyiTpdJXyWB1g2l8jPcqyeWDMqjOnaa9UA=",
        "password": null,
        "passwordHistory": [],
        "passwordRevisionDate": null,
        "totp": "2.mScxVU7uCx3fiPXYlgzWVA==|yxqIjknG71Qdwxq1wyVufWyB1Hb6qWowPgGoZNV2d4s=|uY7hrn7Eow81A76/n4VUrJxSRtC9VDnUdaesO6y0ivY=",
        "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
        "uris": [
          {
            "match": null,
            "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
            "uriChecksum": "2.WfCD+h+rQ49rJDiTM7ycjw==|QgYnnlKzeymjRXmW15SC7bnWWvWaU1iki7rABdi0+RjPgRVGLluT/lQCKXkr5fXH|vaMZaBCKZySaycxVXjpDatKjrN6mBaq2ffRsygLTGzA="
          }
        ],
        "username": "2.9hmypXsnAjVxJ2e5kZQLKA==|8Jif+MQgdTuAlaCn7i/xig==|xy7zJh4qXutESyC02aKSYb/79EmfbGsYlxsYfKqneLA="
      },
      "deletedDate": null,
      "edit": true,
      "favorite": false,
      "fields": [
        {
          "linkedId": null,
          "name": "2.svtF3aK9R1MLHt2GwH7Ykw==|PqVDa02T0Vzx6ogTNz1Xyg==|g9/Gd5OOLkVtyVAk9vkRCY2reWTZjdY1D3XsVpBfQA4=",
          "type": 0,
          "value": "2.K49DvcymvaQ/+V/3soCb1Q==|r+b2dz0YqFeA8BrHii5a0DIjkwxdwlY+aEEH2d2NdC0=|KEJAjqImC0Jit9oaFN2rSZHyetxuD7jEjMt07HPFLvU="
        }
      ],
      "folderId": null,
      "id": "24d1c150-5dfd-4008-964c-01317d1f6b23",
      "identity": null,
      "key": null,
      "login": {
        "autofillOnPageLoad": null,
        "password": null,
        "passwordRevisionDate": null,
        "totp": "2.mScxVU7uCx3fiPXYlgzWVA==|yxqIjknG71Qdwxq1wyVufWyB1Hb6qWowPgGoZNV2d4s=|uY7hrn7Eow81A76/n4VUrJxSRtC9VDnUdaesO6y0ivY=",
        "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
        "uris": [
          {
            "match": null,
            "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
            "uriChecksum": "2.WfCD+h+rQ49rJDiTM7ycjw==|QgYnnlKzeymjRXmW15SC7bnWWvWaU1iki7rABdi0+RjPgRVGLluT/lQCKXkr5fXH|vaMZaBCKZySaycxVXjpDatKjrN6mBaq2ffRsygLTGzA="
          }
        ],
        "username": "2.9hmypXsnAjVxJ2e5kZQLKA==|8Jif+MQgdTuAlaCn7i/xig==|xy7zJh4qXutESyC02aKSYb/79EmfbGsYlxsYfKqneLA="
      },
      "name": "2.LjR8NxRtCB1noDILxNKmPQ==|s4e2AO4I3HqmPRRsbsYl0XYSuWu7+sn2K3+jNiFjsW0=|3NQW64/3RmPxe3hhUGeMTrSy9Ruh5hYRlJxIhhWhhSI=",
      "notes": "2.ke3IFCPe50UCM2XdJDS/VQ==|ksO+dyEuRBgrpAelfhxNhw==|JrbMU58V5QyiTpdJXyWB1g2l8jPcqyeWDMqjOnaa9UA=",
      "object": "cipherDetails",
      "organizationId": null,
      "organizationUseTotp": true,
      "passwordHistory": [],
      "reprompt": 0,
      "revisionDate": "2024-09-22T11:13:40.347356Z",
      "secureNote": null,
      "type": 1,
      "viewPassword": true
    },
    {
      "attachments": null,
      "card": null,
      "collectionIds": ["8d1a5611-5fd6-4728-a6f8-22bc03b50640"],
      "creationDate": "2024-09-22T11:15:49.885670Z",
      "data": {
        "autofillOnPageLoad": null,
        "fields": [],
        "name": "2.kOK8qkac+p+c8mNjhkuO/w==|nP1hQA1/8rJt97NmWSarfchRF+XWAUflsYzq9X2lBPU=|J7AmiEiDRCWvlLjYBrwXxCgCLw4EQiXmvlKnJw0mTb8=",
        "notes": null,
        "password": "2.uiCrgwCc32t6jIaUcMiinw==|JJVfmpcdrekNqJLfJOVaGw==|1zpwpqlATK+6/av/6JfOoHJaOgtpNJA6LDqEqsR2C/A=",
        "passwordHistory": [],
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": "2.6WVjMRz5qtfVkGJ40Ne+Fw==|LtPQSeBvHIEgmzTtpdq4lw==|BV/F2ZhCD1nsr/4RoETwpMnu8PPVYirxrDQXcF6gpg4="
      },
      "deletedDate": null,
      "edit": true,
      "favorite": false,
      "fields": [],
      "folderId": null,
      "id": "94b0e53d-a194-493f-ad25-e6ca9b9abb75",
      "identity": null,
      "key": null,
      "login": {
        "autofillOnPageLoad": null,
        "password": "2.uiCrgwCc32t6jIaUcMiinw==|JJVfmpcdrekNqJLfJOVaGw==|1zpwpqlATK+6/av/6JfOoHJaOgtpNJA6LDqEqsR2C/A=",
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": "2.6WVjMRz5qtfVkGJ40Ne+Fw==|LtPQSeBvHIEgmzTtpdq4lw==|BV/F2ZhCD1nsr/4RoETwpMnu8PPVYirxrDQXcF6gpg4="
      },
      "name": "2.kOK8qkac+p+c8mNjhkuO/w==|nP1hQA1/8rJt97NmWSarfchRF+XWAUflsYzq9X2lBPU=|J7AmiEiDRCWvlLjYBrwXxCgCLw4EQiXmvlKnJw0mTb8=",
      "notes": null,
      "object": "cipherDetails",
      "organizationId": "81cc1652-dc80-472d-909f-9539d057068b",
      "organizationUseTotp": true,
      "passwordHistory": [],
      "reprompt": 0,
      "revisionDate": "2024-09-22T11:15:49.887811Z",
      "secureNote": null,
      "type": 1,
      "viewPassword": true
    },
    {
      "attachments": null,
      "card": null,
      "collectionIds": ["15e18629-849f-4629-895b-54600c542a70"],
      "creationDate": "2024-09-22T11:14:53.442012Z",
      "data": {
        "autofillOnPageLoad": null,
        "fields": [],
        "name": "2.sOEzwZwic+bzHlaUpnl43Q==|Hr5DUOPmBJiz1oT8o2uJNLKFhY8rcmsx8EGqxuoJC9tNBCv6WwJOotrr2bDoPPP/|ZgcnsODyuV2Zd+Nl7MGoAPj383d/fr8LCEIvidYFJYk=",
        "notes": null,
        "password": "2.Nw6WShTt6PTNzRvnLYp2oQ==|x2mPKtOtqPpJEbCPH4EBBw==|Ij7Gq+CCXpJgoZISsATzsZb7v7paSj8mryWKyU0t/VY=",
        "passwordHistory": [],
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": "2.K7I2dQ8x82rPK2buV1dzUQ==|VdjhkLm2wlURH03+GnLTFw==|AtpKj+TYgu+/YSqPn2jkFzePKvz71JtCgn+w0MXuhoI="
      },
      "deletedDate": null,
      "edit": true,
      "favorite": false,
      "fields": [],
      "folderId": null,
      "id": "adaef652-16e0-439a-9d67-07671ddc2a51",
      "identity": null,
      "key": null,
      "login": {
        "autofillOnPageLoad": null,
        "password": "2.Nw6WShTt6PTNzRvnLYp2oQ==|x2mPKtOtqPpJEbCPH4EBBw==|Ij7Gq+CCXpJgoZISsATzsZb7v7paSj8mryWKyU0t/VY=",
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": "2.K7I2dQ8x82rPK2buV1dzUQ==|VdjhkLm2wlURH03+GnLTFw==|AtpKj+TYgu+/YSqPn2jkFzePKvz71JtCgn+w0MXuhoI="
      },
      "name": "2.sOEzwZwic+bzHlaUpnl43Q==|Hr5DUOPmBJiz1oT8o2uJNLKFhY8rcmsx8EGqxuoJC9tNBCv6WwJOotrr2bDoPPP/|ZgcnsODyuV2Zd+Nl7MGoAPj383d/fr8LCEIvidYFJYk=",
      "notes": null,
      "object": "cipherDetails",
      "organizationId": "81cc1652-dc80-472d-909f-9539d057068b",
      "organizationUseTotp": true,
      "passwordHistory": [],
      "reprompt": 0,
      "revisionDate": "2024-09-22T11:14:53.443032Z",
      "secureNote": null,
      "type": 1,
      "viewPassword": true
    },
    {
      "attachments": null,
      "card": null,
      "collectionIds": [],
      "creationDate": "2024-09-26T14:26:27.436093Z",
      "data": {
        "autofillOnPageLoad": null,
        "fields": [
          {
            "linkedId": null,
            "name": null,
            "type": 2,
            "value": "2.5ndrC0WqMJO6fH/2xv3AMw==|zY1/7UgW/MQUutZBixlyOA==|M4/tHU/s5w6ZygyXVUxpiEtAlFtoa9PlL5DY+S8K6gY="
          }
        ],
        "name": "2.4lqddV7PV1Uxn0qFk0JQLA==|svoCTgXohcpPJZirBg1PcQ==|AFIutATBeBvoGpfvDMqaDGlQXBFvIBKTyZqJNt+UPnE=",
        "notes": null,
        "password": null,
        "passwordHistory": [],
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": "2.DkVdwPK5E5RAn6BLwzHXIg==|484sbAo5PFn6D50K9AToYw==|wjCnCnR+JunkRnoRjC/jwcJizdHCuzJ2RRp1z2uuJBg="
      },
      "deletedDate": null,
      "edit": true,
      "favorite": false,
      "fields": [
        {
          "linkedId": null,
          "name": null,
          "type": 2,
          "value": "2.5ndrC0WqMJO6fH/2xv3AMw==|zY1/7UgW/MQUutZBixlyOA==|M4/tHU/s5w6ZygyXVUxpiEtAlFtoa9PlL5DY+S8K6gY="
        }
      ],
      "folderId": "3df04ed6-624d-45ae-b825-7ceb0cf6eb0e",
      "id": "b4b2d9ad-1d86-41d7-a1f3-a249e9f96e8e",
      "identity": null,
      "key": null,
      "login": {
        "autofillOnPageLoad": null,
        "password": null,
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": "2.DkVdwPK5E5RAn6BLwzHXIg==|484sbAo5PFn6D50K9AToYw==|wjCnCnR+JunkRnoRjC/jwcJizdHCuzJ2RRp1z2uuJBg="
      },
      "name": "2.4lqddV7PV1Uxn0qFk0JQLA==|svoCTgXohcpPJZirBg1PcQ==|AFIutATBeBvoGpfvDMqaDGlQXBFvIBKTyZqJNt+UPnE=",
      "notes": null,
      "object": "cipherDetails",
      "organizationId": null,
      "organizationUseTotp": true,
      "passwordHistory": [],
      "reprompt": 0,
      "revisionDate": "2024-09-26T14:26:27.436534Z",
      "secureNote": null,
      "type": 1,
      "viewPassword": true
    },
    {
      "attachments": null,
      "card": null,
      "collectionIds": [],
      "creationDate": "2024-09-22T11:14:14.671495Z",
      "data": {
        "autofillOnPageLoad": null,
        "fields": [],
        "name": "2.aavpFIxp7xouEaflFBr0ug==|JEyMdNkR/1TaJ5lU5w8FfR4+yfV+JP+XsJd8np7/gY4=|BaHDWGBVNTdd5KsR2VLPrHA0caVgxszWFEwgmtR5iHI=",
        "notes": null,
        "password": "2.qjgHReWyJ0X9KlP3qFtB8w==|cHxcCMM2OTTdI3mt85KveQ==|/EuUwk9AzkUQqsJzJ3Z2MEvM4tQ9Ma8WDoUAXtNLE4E=",
        "passwordHistory": [],
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": "2.K69oGBEuH9GzhSkzi2zzIg==|bbeKWbkuSc1cVJm2xDae8Q==|y+Z07fAT+zmlq2gSLFC6uLEfYIq3IT7HqNFJqlWQOPM="
      },
      "deletedDate": null,
      "edit": true,
      "favorite": false,
      "fields": [],
      "folderId": "e7098f5d-4d00-4bdf-8c66-56f9dca2129f",
      "id": "b7441ea3-baf4-4f04-a445-0ad342eb56c4",
      "identity": null,
      "key": null,
      "login": {
        "autofillOnPageLoad": null,
        "password": "2.qjgHReWyJ0X9KlP3qFtB8w==|cHxcCMM2OTTdI3mt85KveQ==|/EuUwk9AzkUQqsJzJ3Z2MEvM4tQ9Ma8WDoUAXtNLE4E=",
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": "2.K69oGBEuH9GzhSkzi2zzIg==|bbeKWbkuSc1cVJm2xDae8Q==|y+Z07fAT+zmlq2gSLFC6uLEfYIq3IT7HqNFJqlWQOPM="
      },
      "name": "2.aavpFIxp7xouEaflFBr0ug==|JEyMdNkR/1TaJ5lU5w8FfR4+yfV+JP+XsJd8np7/gY4=|BaHDWGBVNTdd5KsR2VLPrHA0caVgxszWFEwgmtR5iHI=",
      "notes": null,
      "object": "cipherDetails",
      "organizationId": null,
      "organizationUseTotp": true,
      "passwordHistory": [],
      "reprompt": 0,
      "revisionDate": "2024-09-22T11:14:14.672096Z",
      "secureNote": null,
      "type": 1,
      "viewPassword": true
    }
  ],
  "collections": [
    {
      "externalId": null,
      "hidePasswords": false,
      "id": "15e18629-849f-4629-895b-54600c542a70",
      "name": "2.X7RK7wBZl+1pqo4yvAontQ==|Q+RD7xRWtXPC/Ijbl+863VSyQ4oYmHGG5EzNDtNYWW0=|369crUvCUpo5bID5wu0Q60eQJJv5+eJLxbsuZKmtVYA=",
      "object": "collectionDetails",
      "organizationId": "81cc1652-dc80-472d-909f-9539d057068b",
      "readOnly": false
    },
    {
      "externalId": null,
      "hidePasswords": false,
      "id": "8d1a5611-5fd6-4728-a6f8-22bc03b50640",
      "name": "2.4KNb9dUNUqfe2iqFZxBZJQ==|lrcBRuOjIpMAiBMHYDNbpn6ZmUz/OzHuBTr68ta0a/I=|OtEMhyhekHnARLATP8IWKAWieTaNQpbgO7/syW1KLK4=",
      "object": "collectionDetails",
      "organizationId": "81cc1652-dc80-472d-909f-9539d057068b",
      "readOnly": false
    }
  ],
  "domains": null,
  "folders": [
    {
      "id": "e7098f5d-4d00-4bdf-8c66-56f9dca2129f",
      "name": "2.FTcL7PAe36ZQh0zIXPotSg==|ALhJRf/j9Wa83jg/tWbFSndiZ9JvZs2BwmVXmCUWhKc=|AEpDm/0LhllnW/Qzx9oWNYuSMaLQuR62MFUQuAaqQvc=",
      "object": "folder",
      "revisionDate": "2024-09-22T11:13:55.612053Z"
    },
    {
      "id": "3df04ed6-624d-45ae-b825-7ceb0cf6eb0e",
      "name": "2.cybmeku3f53n3//scS4HfQ==|QEOsVN2jUF3+11/a8iISuQ==|PV2zCoVqNCX9KzyIUassEOjZE/f6bF+Y6Ft1ypRJ9lw=",
      "object": "folder",
      "revisionDate": "2024-09-22T17:00:11.566430Z"
    }
  ],
  "object": "sync",
  "policies": [],
  "profile": {
    "_status": 0,
    "avatarColor": null,
    "culture": "en-US",
    "email": "test@laverse.net",
    "emailVerified": true,
    "forcePasswordReset": false,
    "id": "e8dababd-242e-4900-becf-e88bc021dda8",
    "key": "2.lkAJiJtCKPHFPrZ96+j2Xg==|5XJtrKUndcGy28thFukrmgMcLp+BOVdkF+KcuOnfshq9AN1PFhna9Es96CVARCnjTcWuHuqvgnGmcOHTrf8fyfLv63VBsjLgLZk8rCXJoKE=|9dwgx4/13AD+elE2vE7vlSQoe8LbCGGlui345YrKvXY=",
    "masterPasswordHint": null,
    "name": "Test Laverse",
    "object": "profile",
    "organizations": [
      {
        "accessSecretsManager": false,
        "allowAdminAccessToAllCollectionItems": true,
        "enabled": true,
        "familySponsorshipAvailable": false,
        "familySponsorshipFriendlyName": null,
        "familySponsorshipLastSyncDate": null,
        "familySponsorshipToDelete": null,
        "familySponsorshipValidUntil": null,
        "flexibleCollections": false,
        "hasPublicAndPrivateKeys": true,
        "id": "81cc1652-dc80-472d-909f-9539d057068b",
        "identifier": null,
        "key": "4.JW3mktbL7vpTVRweZdQBAirJuAEhSRn37zcXZjDjI47weFKkeZkvPxZWCqFYC/P5qCJwEYMbv7lTETkWDg6paevVfhJ35buGcTQdEbQxAJebzPahEcUstj11l4Y9T5RaDiAJR8+drrGJ3fKV3v3hymKz2o9fUfK1epuLFll2nnWSOjCcuRe/+zz5VwIVx4WJAPJHmiS6eofbj/DTIQCzG4JkR0UzT66ouLcgmPL1nGOqVI7KxRpL5yVj75UkjniHkWAcB7lfAxWXw2GhDJ/2L685uA3820ItTbxjCwLQOvjBttgrbURmkeP9BD+KkO4V6vb8bbTWNSvggXKk2h1CMw==",
        "keyConnectorEnabled": false,
        "keyConnectorUrl": null,
        "limitCollectionCreationDeletion": true,
        "maxAutoscaleSeats": null,
        "maxCollections": null,
        "maxStorageGb": 32767,
        "name": "My Test Organization",
        "object": "profileOrganization",
        "organizationUserId": "9e755d16-b500-48ac-b24c-4105b7d08796",
        "permissions": {
          "accessEventLogs": false,
          "accessImportExport": false,
          "accessReports": false,
          "createNewCollections": false,
          "deleteAnyCollection": false,
          "deleteAssignedCollections": false,
          "editAnyCollection": false,
          "editAssignedCollections": false,
          "manageGroups": false,
          "managePolicies": false,
          "manageResetPassword": false,
          "manageScim": false,
          "manageSso": false,
          "manageUsers": false
        },
        "planProductType": 3,
        "productTierType": 3,
        "providerId": null,
        "providerName": null,
        "providerType": null,
        "resetPasswordEnrolled": false,
        "seats": null,
        "selfHost": true,
        "ssoBound": false,
        "status": 2,
        "type": 0,
        "use2fa": true,
        "useActivateAutofillPolicy": false,
        "useApi": true,
        "useCustomPermissions": false,
        "useDirectory": false,
        "useEvents": false,
        "useGroups": false,
        "useKeyConnector": false,
        "usePasswordManager": true,
        "usePolicies": true,
        "useResetPassword": false,
        "useScim": false,
        "useSecretsManager": false,
        "useSso": false,
        "useTotp": true,
        "userId": "e8dababd-242e-4900-becf-e88bc021dda8",
        "usersGetPremium": true
      }
    ],
    "premium": true,
    "premiumFromOrganization": false,
    "privateKey": "2.D2aLa8ne/DAkeSzctQISVw==|/xoGM5i5JGJTH/vohUuwTFrTx3hd/gt3kBD/FQdeLMtYjl1u96sh0ECmoERqGHeSfXj+iAb9kpTIOKG8LwmkZGUJBI90Mw0M7ODmf7E8eQ+aGF+bGqTSMQ1wtpunEyFVodlg92YN8Ddlb2V9J4uN8ykpHYNDmQiYLZ8bl6vCODRGPyzLvx5M8DbITVL5PhsjKDLLrVV8lFCgCcAL5YLfkghYhFELyX15zXA/KYEnwggDka3hG5+HHFOVZSeyk7Gi6M4TX2wADbTXz1/Wsho8oxFUrtNiOB3ZiY2cx9UWttpzMXoGfi2gJcP1db/nTfWenOLlzw6Od4VyRzsXsfyGwbqBqDnNFkjLvhjVw4JO+psF//xAMDs14101Tf2wFkB6toQ+zdnDphXUeKmiVPQ7gMnQlOWN5tWvjjmYOO4Y63sGpP24cDOdEScIdebZRSA8uOhTzadfKfOiH5zVYZzXs33FQ0li4nBrsj5xYa6PP4D1P7gqjxClPdguwkdLoZ7JvgIlyRIwEcORi5Ich8RWF/kqRBwk0QSzK1mTlHHU5xtSgi4MLNVx4qTYNaJVBtwL9d2MD9LeNn4Z2PL4A7qnszHqsERiQQDxNEgMxMBHDgSXqQbtQxRvsI6oY+yNbN7uVWw4o8AC3f5GBdxzIqcN1mgEM5ix5aDt15w3MhP2FHtf2neKI68TnL8WnT1fT8BVlbECIiUqK0tfq5aTjdSh3gCS15jvZ1H60h+K8+O/nDfquzVjY7UsTGwA+UtS8/JGiaUhc0VhxJo8P1V2VSCiu51d5q3De1vDg5R2VEgBmTchZyTIodC+3+7ACTOwkNCCdIJN7xKOcIGFA7QOuyJtBeXT4Rd9UGMHSL054IB/315WVDiwrP9W1aP0nHzFs+qAXbH5o1E+AmfMDyoHopjGgbUw22r837kHzf5Qe8QhRYPzQvDowfCPhdoy23cbN1VsNavNwTC9hcG5oYMwkK66xP3MEM9UfGD22pwxe7M8U4BRdLCCHbi95eklXE6Mg6DpWsAdMgokQbOvnlwgKfrlbltqXUE/vUQI8TB3AE1Nkt0ST4quTriMuuyiHdeeZV4UkV9jWt/5bTCfdrCYuGZP7g4shbfNcP7u1Zrdxv+EuUwGIOOTrNV5awmBnL3iHE5ya2MnqmRyfWiPIT5majZCk06yxj4XzyIPOpjYKFt0MOgLvG1GllmdtRqg7tMVvc5ZFo5KWIxLsIJD12UjA1GYYoFdX4+wsNbPjfnlE6D2PrtWUICnBFJzYpfyrKTe01k8G8hyyz+tVzBRfz8EA2ew1+hlVcAgSPCcBzhDgqPe+RSPi7ZSd66be1gDhGAftWFM8Z0MrMklXi2DyjjaKBNsZZD8qTcLcobm8nqHUQtnr5JCbmgP3rau8NY/fxeFHsvSiZQoB1aI/y+Sz/R4r+T9cg8hjmS/FUHDO+m6a6nuWNFwz8wIluM557oOTl+A9UGFF50Gpzmf97VdQjM3ZREazQ7la6AobzS3BHI6FNdxN9LTyMpYo+WODv52/VwU3ODH7wf5bz2OHZhk2NG5R7pSH7qg8jM+/MtJkFumENV0qMecozIkP6e4CyI9ua4YwI9n7G5OgKYMG1aj2PRSny2JSLS8aHF1TkRL8SD0nZFCox0=|muEtiwIuZxhuuLv0nouEdxHU2CO+I7JXKZuYHWiv/OE=",
    "providerOrganizations": [],
    "providers": [],
    "securityStamp": "4033cf35-bc5d-4ecd-951c-dbe549f27740",
    "twoFactorEnabled": false,
    "usesKeyConnector": false
  },
  "sends": [],
  "unofficialServer": true
}
`))

	httpmock.RegisterResponder("GET", `http://127.0.0.1:8081/api/accounts/profile`,
		httpmock.NewStringResponder(200, `{"_status":0,"avatarColor":null,"culture":"en-US","email":"test@laverse.net","emailVerified":true,"forcePasswordReset":false,"id":"e8dababd-242e-4900-becf-e88bc021dda8","key":"2.lkAJiJtCKPHFPrZ96+j2Xg==|5XJtrKUndcGy28thFukrmgMcLp+BOVdkF+KcuOnfshq9AN1PFhna9Es96CVARCnjTcWuHuqvgnGmcOHTrf8fyfLv63VBsjLgLZk8rCXJoKE=|9dwgx4/13AD+elE2vE7vlSQoe8LbCGGlui345YrKvXY=","masterPasswordHint":null,"name":"Test Laverse","object":"profile","organizations":[{"accessSecretsManager":false,"allowAdminAccessToAllCollectionItems":true,"enabled":true,"familySponsorshipAvailable":false,"familySponsorshipFriendlyName":null,"familySponsorshipLastSyncDate":null,"familySponsorshipToDelete":null,"familySponsorshipValidUntil":null,"flexibleCollections":false,"hasPublicAndPrivateKeys":true,"id":"81cc1652-dc80-472d-909f-9539d057068b","identifier":null,"key":"4.JW3mktbL7vpTVRweZdQBAirJuAEhSRn37zcXZjDjI47weFKkeZkvPxZWCqFYC/P5qCJwEYMbv7lTETkWDg6paevVfhJ35buGcTQdEbQxAJebzPahEcUstj11l4Y9T5RaDiAJR8+drrGJ3fKV3v3hymKz2o9fUfK1epuLFll2nnWSOjCcuRe/+zz5VwIVx4WJAPJHmiS6eofbj/DTIQCzG4JkR0UzT66ouLcgmPL1nGOqVI7KxRpL5yVj75UkjniHkWAcB7lfAxWXw2GhDJ/2L685uA3820ItTbxjCwLQOvjBttgrbURmkeP9BD+KkO4V6vb8bbTWNSvggXKk2h1CMw==","keyConnectorEnabled":false,"keyConnectorUrl":null,"limitCollectionCreationDeletion":true,"maxAutoscaleSeats":null,"maxCollections":null,"maxStorageGb":32767,"name":"My Test Organization","object":"profileOrganization","organizationUserId":"9e755d16-b500-48ac-b24c-4105b7d08796","permissions":{"accessEventLogs":false,"accessImportExport":false,"accessReports":false,"createNewCollections":false,"deleteAnyCollection":false,"deleteAssignedCollections":false,"editAnyCollection":false,"editAssignedCollections":false,"manageGroups":false,"managePolicies":false,"manageResetPassword":false,"manageScim":false,"manageSso":false,"manageUsers":false},"planProductType":3,"productTierType":3,"providerId":null,"providerName":null,"providerType":null,"resetPasswordEnrolled":false,"seats":null,"selfHost":true,"ssoBound":false,"status":2,"type":0,"use2fa":true,"useActivateAutofillPolicy":false,"useApi":true,"useCustomPermissions":false,"useDirectory":false,"useEvents":false,"useGroups":false,"useKeyConnector":false,"usePasswordManager":true,"usePolicies":true,"useResetPassword":false,"useScim":false,"useSecretsManager":false,"useSso":false,"useTotp":true,"userId":"e8dababd-242e-4900-becf-e88bc021dda8","usersGetPremium":true}],"premium":true,"premiumFromOrganization":false,"privateKey":"2.D2aLa8ne/DAkeSzctQISVw==|/xoGM5i5JGJTH/vohUuwTFrTx3hd/gt3kBD/FQdeLMtYjl1u96sh0ECmoERqGHeSfXj+iAb9kpTIOKG8LwmkZGUJBI90Mw0M7ODmf7E8eQ+aGF+bGqTSMQ1wtpunEyFVodlg92YN8Ddlb2V9J4uN8ykpHYNDmQiYLZ8bl6vCODRGPyzLvx5M8DbITVL5PhsjKDLLrVV8lFCgCcAL5YLfkghYhFELyX15zXA/KYEnwggDka3hG5+HHFOVZSeyk7Gi6M4TX2wADbTXz1/Wsho8oxFUrtNiOB3ZiY2cx9UWttpzMXoGfi2gJcP1db/nTfWenOLlzw6Od4VyRzsXsfyGwbqBqDnNFkjLvhjVw4JO+psF//xAMDs14101Tf2wFkB6toQ+zdnDphXUeKmiVPQ7gMnQlOWN5tWvjjmYOO4Y63sGpP24cDOdEScIdebZRSA8uOhTzadfKfOiH5zVYZzXs33FQ0li4nBrsj5xYa6PP4D1P7gqjxClPdguwkdLoZ7JvgIlyRIwEcORi5Ich8RWF/kqRBwk0QSzK1mTlHHU5xtSgi4MLNVx4qTYNaJVBtwL9d2MD9LeNn4Z2PL4A7qnszHqsERiQQDxNEgMxMBHDgSXqQbtQxRvsI6oY+yNbN7uVWw4o8AC3f5GBdxzIqcN1mgEM5ix5aDt15w3MhP2FHtf2neKI68TnL8WnT1fT8BVlbECIiUqK0tfq5aTjdSh3gCS15jvZ1H60h+K8+O/nDfquzVjY7UsTGwA+UtS8/JGiaUhc0VhxJo8P1V2VSCiu51d5q3De1vDg5R2VEgBmTchZyTIodC+3+7ACTOwkNCCdIJN7xKOcIGFA7QOuyJtBeXT4Rd9UGMHSL054IB/315WVDiwrP9W1aP0nHzFs+qAXbH5o1E+AmfMDyoHopjGgbUw22r837kHzf5Qe8QhRYPzQvDowfCPhdoy23cbN1VsNavNwTC9hcG5oYMwkK66xP3MEM9UfGD22pwxe7M8U4BRdLCCHbi95eklXE6Mg6DpWsAdMgokQbOvnlwgKfrlbltqXUE/vUQI8TB3AE1Nkt0ST4quTriMuuyiHdeeZV4UkV9jWt/5bTCfdrCYuGZP7g4shbfNcP7u1Zrdxv+EuUwGIOOTrNV5awmBnL3iHE5ya2MnqmRyfWiPIT5majZCk06yxj4XzyIPOpjYKFt0MOgLvG1GllmdtRqg7tMVvc5ZFo5KWIxLsIJD12UjA1GYYoFdX4+wsNbPjfnlE6D2PrtWUICnBFJzYpfyrKTe01k8G8hyyz+tVzBRfz8EA2ew1+hlVcAgSPCcBzhDgqPe+RSPi7ZSd66be1gDhGAftWFM8Z0MrMklXi2DyjjaKBNsZZD8qTcLcobm8nqHUQtnr5JCbmgP3rau8NY/fxeFHsvSiZQoB1aI/y+Sz/R4r+T9cg8hjmS/FUHDO+m6a6nuWNFwz8wIluM557oOTl+A9UGFF50Gpzmf97VdQjM3ZREazQ7la6AobzS3BHI6FNdxN9LTyMpYo+WODv52/VwU3ODH7wf5bz2OHZhk2NG5R7pSH7qg8jM+/MtJkFumENV0qMecozIkP6e4CyI9ua4YwI9n7G5OgKYMG1aj2PRSny2JSLS8aHF1TkRL8SD0nZFCox0=|muEtiwIuZxhuuLv0nouEdxHU2CO+I7JXKZuYHWiv/OE=","providerOrganizations":[],"providers":[],"securityStamp":"4033cf35-bc5d-4ecd-951c-dbe549f27740","twoFactorEnabled":false,"usesKeyConnector":false}`))

	// We're changing the revisionDate on purpose in the response to mimick Bitwarden official server's behavior
	httpmock.RegisterResponder("POST", `http://127.0.0.1:8081/api/ciphers`,
		httpmock.NewStringResponder(200, `
		{
      "attachments": null,
      "card": null,
      "collectionIds": [],
      "creationDate": "2024-09-22T11:13:40.346903Z",
      "data": {
        "autofillOnPageLoad": null,
        "fields": [
          {
            "linkedId": null,
            "name": "2.svtF3aK9R1MLHt2GwH7Ykw==|PqVDa02T0Vzx6ogTNz1Xyg==|g9/Gd5OOLkVtyVAk9vkRCY2reWTZjdY1D3XsVpBfQA4=",
            "type": 0,
            "value": "2.K49DvcymvaQ/+V/3soCb1Q==|r+b2dz0YqFeA8BrHii5a0DIjkwxdwlY+aEEH2d2NdC0=|KEJAjqImC0Jit9oaFN2rSZHyetxuD7jEjMt07HPFLvU="
          }
        ],
        "name": "2.LjR8NxRtCB1noDILxNKmPQ==|s4e2AO4I3HqmPRRsbsYl0XYSuWu7+sn2K3+jNiFjsW0=|3NQW64/3RmPxe3hhUGeMTrSy9Ruh5hYRlJxIhhWhhSI=",
        "notes": "2.ke3IFCPe50UCM2XdJDS/VQ==|ksO+dyEuRBgrpAelfhxNhw==|JrbMU58V5QyiTpdJXyWB1g2l8jPcqyeWDMqjOnaa9UA=",
        "password": null,
        "passwordHistory": [],
        "passwordRevisionDate": null,
        "totp": "2.mScxVU7uCx3fiPXYlgzWVA==|yxqIjknG71Qdwxq1wyVufWyB1Hb6qWowPgGoZNV2d4s=|uY7hrn7Eow81A76/n4VUrJxSRtC9VDnUdaesO6y0ivY=",
        "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
        "uris": [
          {
            "match": null,
            "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
            "uriChecksum": "2.WfCD+h+rQ49rJDiTM7ycjw==|QgYnnlKzeymjRXmW15SC7bnWWvWaU1iki7rABdi0+RjPgRVGLluT/lQCKXkr5fXH|vaMZaBCKZySaycxVXjpDatKjrN6mBaq2ffRsygLTGzA="
          }
        ],
        "username": "2.9hmypXsnAjVxJ2e5kZQLKA==|8Jif+MQgdTuAlaCn7i/xig==|xy7zJh4qXutESyC02aKSYb/79EmfbGsYlxsYfKqneLA="
      },
      "deletedDate": null,
      "edit": true,
      "favorite": false,
      "fields": [
        {
          "linkedId": null,
          "name": "2.svtF3aK9R1MLHt2GwH7Ykw==|PqVDa02T0Vzx6ogTNz1Xyg==|g9/Gd5OOLkVtyVAk9vkRCY2reWTZjdY1D3XsVpBfQA4=",
          "type": 0,
          "value": "2.K49DvcymvaQ/+V/3soCb1Q==|r+b2dz0YqFeA8BrHii5a0DIjkwxdwlY+aEEH2d2NdC0=|KEJAjqImC0Jit9oaFN2rSZHyetxuD7jEjMt07HPFLvU="
        }
      ],
      "folderId": null,
      "id": "24d1c150-5dfd-4008-964c-01317d1f6b23",
      "identity": null,
      "key": null,
      "login": {
        "autofillOnPageLoad": null,
        "password": null,
        "passwordRevisionDate": null,
        "totp": "2.mScxVU7uCx3fiPXYlgzWVA==|yxqIjknG71Qdwxq1wyVufWyB1Hb6qWowPgGoZNV2d4s=|uY7hrn7Eow81A76/n4VUrJxSRtC9VDnUdaesO6y0ivY=",
        "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
        "uris": [
          {
            "match": null,
            "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
            "uriChecksum": "2.WfCD+h+rQ49rJDiTM7ycjw==|QgYnnlKzeymjRXmW15SC7bnWWvWaU1iki7rABdi0+RjPgRVGLluT/lQCKXkr5fXH|vaMZaBCKZySaycxVXjpDatKjrN6mBaq2ffRsygLTGzA="
          }
        ],
        "username": "2.9hmypXsnAjVxJ2e5kZQLKA==|8Jif+MQgdTuAlaCn7i/xig==|xy7zJh4qXutESyC02aKSYb/79EmfbGsYlxsYfKqneLA="
      },
      "name": "2.LjR8NxRtCB1noDILxNKmPQ==|s4e2AO4I3HqmPRRsbsYl0XYSuWu7+sn2K3+jNiFjsW0=|3NQW64/3RmPxe3hhUGeMTrSy9Ruh5hYRlJxIhhWhhSI=",
      "notes": "2.ke3IFCPe50UCM2XdJDS/VQ==|ksO+dyEuRBgrpAelfhxNhw==|JrbMU58V5QyiTpdJXyWB1g2l8jPcqyeWDMqjOnaa9UA=",
      "object": "cipherDetails",
      "organizationId": null,
      "organizationUseTotp": true,
      "passwordHistory": [],
      "reprompt": 0,
      "revisionDate": "2024-09-22T11:13:40.345356Z",
      "secureNote": null,
      "type": 1,
      "viewPassword": true
    }`))

	return webapi.NewClient("http://127.0.0.1:8081/", webapi.WithCustomClient(client), webapi.DisableRetries())
}

func mockedClientArgon2() webapi.Client {
	client := http.Client{Transport: httpmock.DefaultTransport}
	httpmock.RegisterResponder("POST", "http://127.0.0.1:8081/identity/accounts/prelogin",
		httpmock.NewStringResponder(200, `{"kdf":1,"kdfIterations":3,"kdfMemory":64,"kdfParallelism":4}`))

	// Regexp match (could use httpmock.RegisterRegexpResponder instead)
	httpmock.RegisterResponder("POST", `http://127.0.0.1:8081/identity/connect/token`,
		httpmock.NewStringResponder(200, `{"ForcePasswordReset":false,"Kdf":1,"KdfIterations":3,"KdfMemory":64,"KdfParallelism":4,"Key":"2.VBpAJrIHYLv60UTYy5e7sA==|WWsozKnKPVvBYttXtlAzpmZPoffwlY3+8Wup9SBdGcO8T4Ybj808DubDhICTPMz9RliDSJ1OuaivBC0rh/7EcWV1s9KuRth5J3XFWJqmtXU=|l1ly+Uek1j4k2xQNT8iJ++GdwMTWRaSBP5mFvgJ+CGU=","MasterPasswordPolicy":{"object":"masterPasswordPolicy"},"PrivateKey":"2.0Xc+uV+6YVa+Zkg74C/MJA==|xTbZGy2r8/kHcXXy23YDHWDTHjasUUoHiWJ84rGEhuPD27GTF8DrFL+lbERo3+OH7MXvseAFygKArvSumlCQnMhT/dswJMH8ZEMuxQUZ8I7kg/7eNqxRndxyC8alZh3VM4FC7TsuVBe6BF4pbPQYX4etFL8yOHLpIhwkKcG7+IfQHQHwOXpsSiGpADQlk5er990snhAXwGpgROqoveO9klpNJXuEzJKMFMdoo9FCaRsJ4bB3BE/Y+8Ph7ek8mGyoUUqNFp8Z/T+XN+9kvxDFnVFtYp1p+sU8LpCSvaq3dAImQt6X7vAgfjbWVkTxB0HlBdMkjpg8BAK7qpscT5oH83SZqdNPRi9kkT9Y30xl0QvvJzXLcjqOS6je+i3vB1r4O8X3ZIj+th4cfZOfHKpzDLbGyAezxiiWwQ4xt0USQ9rLv+BU5YNDKnibmsIooKHuh8wV0qht8vr2CEVavXOmmBi/6bGIWJMOs2K52be5LJkYL+653dsKXBN4uahQDMdfs9vPUA7OoIFR9BZvQiGMDstFYVgJUulOkj2J4nieuTAmurQe6nS6U4v1swI7DIzdo8ZCJcAQfWYWk8IBwh6gQxSuPMg7+O2dntbPFkMP2KhoisUucoXNIDXmxwYAMeP01KMNSgFc6hL6WnLzyVcRlSLy7OC9qwTD6ZH6XA4D/MO9MINTTzw4+/pZvwTeuXMqB0HkcjTpWzTUi405uV6qh8CxYOaNEl6kNAMiuVNASToqYb/EBC0uGOybOzcNZOCiXmuiPVOstiz65KP1d3Dfl9al48hpxnDxFyFub5NDx8xADA3SzOlI9xMCsC/mcY0/fRlwoTDFCLzfdtPIdSn5N8/YySY10e/TXftKGV7bLvzGAlOntwoiWaLJbyHAnEXNUVFxJu5XVaVccXcxIKete42S291dCu4yk35KsIQU0jBaPB/hGXLJvvySl9/kARl0zIXJH+Pk0hlk+/IRH9HNZrru93WTgW9KjEeT2vaZ6UqULkkspIfoUrFxQfSyAxycDaqa4EHt1QJBKyC9+aWEOXNwtxkJTnhtvlqRPOGUpdkAyC5ebfdX2URZnd/2TR/PTviWaDKe/g51UWr8QepgwbJjkBeuZMSPCNHkcLdmwEZERisbZ7H50hDQhhK+qfmMOvkVrRRXKT2ICbQzchm5rDasdtMed7vIf3beS0ESIR8kZ1WmteI9dvWmagnXRlfsUZW0y0KWe9Ma/ISOa6QPb0Cc7T8LUg2lpREQpRTt119RX9Xugd2uZ+UzbGpqn9u5JY2+vddwv+zFL0g/Vu9h3kd02v+WTXeV8wLDJYkoc6XeTBzC0uQtNJt4oN7hWcMBrkYmnXxslIjGw/nE7eBnG02XWTCeTjADqXtz2j4xMFmzJod+4j4f6MQZBII9Sz5A2FfSbadUuukEYcFvkqLnjG/sMzcBQMIDP6vmy7VRnETahlNi7Pt+gRLHuhGN1BbScolW7a0YnWlI1T0MtKb4DBsjX9GfotxkG+ORg/i45YnG6KPQ6+jgXU3Xtf0tgjTqHVgc7TOCm9R2b4G24RSQXx3WaMP1slbSYsf/YmXIqdg/dVaCFrJdA1v2bDFymuPbA0P4Ey8c4ylInXGgK4yxS7nVC4qbbVtK5bWT3SRaBWw=|0r9TcWfl93tZln8+ZGrPwZTHBbrLiSuVXvoI+cQeSr0=","ResetMasterPassword":false,"UserDecryptionOptions":{"HasMasterPassword":true,"Object":"userDecryptionOptions"},"access_token":"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJuYmYiOjE3Mjc4MDcxNDUsImV4cCI6MTcyNzgxNDM0NSwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdHxsb2dpbiIsInN1YiI6ImM2NGExZjhkLTIyM2MtNDBmZC1hZGRlLWE5MGQ1OGRiYzI2YyIsInByZW1pdW0iOnRydWUsIm5hbWUiOiJ0ZXN0LTIwMjM0MSIsImVtYWlsIjoidGVzdC0yMDIzNDEtYXJnb24yQGxhdmVyc2UubmV0IiwiZW1haWxfdmVyaWZpZWQiOnRydWUsInNzdGFtcCI6IjZjYmY3Y2QxLWQyYTItNGM1YS05N2Q3LTkxMDQyMjg1YzQ5ZCIsImRldmljZSI6IiIsInNjb3BlIjpbImFwaSIsIm9mZmxpbmVfYWNjZXNzIl0sImFtciI6WyJBcHBsaWNhdGlvbiJdfQ.hJ5-aPuZFAovG_2q7im8xxO2GCjeHit7Hwhm-FKQsu-j7OJdY8F_2jSy7azPcrVasImboR99WvpydMYzbQLB3AlQ3OSPcqUUHJDlFqeDUIXzpi-NMb3CBkxyr8Q_fBojiWmRiwPY89oiOE6dXTvDxTEUIWpKv0xnd-iBBPwwx6vzfsSKKAIB1tbZM1Cy4aWgcbN0lC-cU-NGcIFI1UEVod9KzcuN65DouXfjB2d5NkAI-pic9pno42MF8strosezgWrbnvTk8h7CzT5BPj7ZCaqYVr_RKVciHAs3QwrFyFe_OZ4RoN_YikDApdbwQyIsJsIK_jNTraJzNMfo6bbZfA","expires_in":7200,"refresh_token":"JPxgcZSlj4KKfgZuyf8PnLF9KVXADYOYstA5hBreXGPwm7667dSeGPFtBYRWOyXRJxQryUr6BLEjXHBS_NqdIw==","scope":"api offline_access","token_type":"Bearer","unofficialServer":true}`))

	httpmock.RegisterResponder("GET", `http://127.0.0.1:8081/api/sync?excludeDomains=true`,
		httpmock.NewStringResponder(200, `{
  "ciphers": [
    {
      "attachments": null,
      "card": null,
      "collectionIds": [],
      "creationDate": "2024-10-01T18:25:26.092790Z",
      "data": {
        "autofillOnPageLoad": null,
        "fields": [],
        "name": "2.zD1jZaHjvJ0ZsD4OpfRHXg==|eekdaveWXtPM+Bo285ljIQ==|4pvvvQZzyX5HRCaY+jUrecLR0AlXx4oVf994NNTyHyE=",
        "notes": null,
        "password": null,
        "passwordHistory": [],
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": null
      },
      "deletedDate": null,
      "edit": true,
      "favorite": false,
      "fields": [],
      "folderId": null,
      "id": "f01e68e9-3951-40ee-80ff-c4ff4517e159",
      "identity": null,
      "key": null,
      "login": {
        "autofillOnPageLoad": null,
        "password": null,
        "passwordRevisionDate": null,
        "totp": null,
        "uri": null,
        "uris": [],
        "username": null
      },
      "name": "2.zD1jZaHjvJ0ZsD4OpfRHXg==|eekdaveWXtPM+Bo285ljIQ==|4pvvvQZzyX5HRCaY+jUrecLR0AlXx4oVf994NNTyHyE=",
      "notes": null,
      "object": "cipherDetails",
      "organizationId": null,
      "organizationUseTotp": true,
      "passwordHistory": [],
      "reprompt": 0,
      "revisionDate": "2024-10-01T18:25:26.093577Z",
      "secureNote": null,
      "type": 1,
      "viewPassword": true
    }
  ],
  "collections": [],
  "domains": null,
  "folders": [],
  "object": "sync",
  "policies": [],
  "profile": {
    "_status": 0,
    "avatarColor": null,
    "culture": "en-US",
    "email": "test-202341-argon2@laverse.net",
    "emailVerified": true,
    "forcePasswordReset": false,
    "id": "c64a1f8d-223c-40fd-adde-a90d58dbc26c",
    "key": "2.VBpAJrIHYLv60UTYy5e7sA==|WWsozKnKPVvBYttXtlAzpmZPoffwlY3+8Wup9SBdGcO8T4Ybj808DubDhICTPMz9RliDSJ1OuaivBC0rh/7EcWV1s9KuRth5J3XFWJqmtXU=|l1ly+Uek1j4k2xQNT8iJ++GdwMTWRaSBP5mFvgJ+CGU=",
    "masterPasswordHint": null,
    "name": "test-202341",
    "object": "profile",
    "organizations": [],
    "premium": true,
    "premiumFromOrganization": false,
    "privateKey": "2.0Xc+uV+6YVa+Zkg74C/MJA==|xTbZGy2r8/kHcXXy23YDHWDTHjasUUoHiWJ84rGEhuPD27GTF8DrFL+lbERo3+OH7MXvseAFygKArvSumlCQnMhT/dswJMH8ZEMuxQUZ8I7kg/7eNqxRndxyC8alZh3VM4FC7TsuVBe6BF4pbPQYX4etFL8yOHLpIhwkKcG7+IfQHQHwOXpsSiGpADQlk5er990snhAXwGpgROqoveO9klpNJXuEzJKMFMdoo9FCaRsJ4bB3BE/Y+8Ph7ek8mGyoUUqNFp8Z/T+XN+9kvxDFnVFtYp1p+sU8LpCSvaq3dAImQt6X7vAgfjbWVkTxB0HlBdMkjpg8BAK7qpscT5oH83SZqdNPRi9kkT9Y30xl0QvvJzXLcjqOS6je+i3vB1r4O8X3ZIj+th4cfZOfHKpzDLbGyAezxiiWwQ4xt0USQ9rLv+BU5YNDKnibmsIooKHuh8wV0qht8vr2CEVavXOmmBi/6bGIWJMOs2K52be5LJkYL+653dsKXBN4uahQDMdfs9vPUA7OoIFR9BZvQiGMDstFYVgJUulOkj2J4nieuTAmurQe6nS6U4v1swI7DIzdo8ZCJcAQfWYWk8IBwh6gQxSuPMg7+O2dntbPFkMP2KhoisUucoXNIDXmxwYAMeP01KMNSgFc6hL6WnLzyVcRlSLy7OC9qwTD6ZH6XA4D/MO9MINTTzw4+/pZvwTeuXMqB0HkcjTpWzTUi405uV6qh8CxYOaNEl6kNAMiuVNASToqYb/EBC0uGOybOzcNZOCiXmuiPVOstiz65KP1d3Dfl9al48hpxnDxFyFub5NDx8xADA3SzOlI9xMCsC/mcY0/fRlwoTDFCLzfdtPIdSn5N8/YySY10e/TXftKGV7bLvzGAlOntwoiWaLJbyHAnEXNUVFxJu5XVaVccXcxIKete42S291dCu4yk35KsIQU0jBaPB/hGXLJvvySl9/kARl0zIXJH+Pk0hlk+/IRH9HNZrru93WTgW9KjEeT2vaZ6UqULkkspIfoUrFxQfSyAxycDaqa4EHt1QJBKyC9+aWEOXNwtxkJTnhtvlqRPOGUpdkAyC5ebfdX2URZnd/2TR/PTviWaDKe/g51UWr8QepgwbJjkBeuZMSPCNHkcLdmwEZERisbZ7H50hDQhhK+qfmMOvkVrRRXKT2ICbQzchm5rDasdtMed7vIf3beS0ESIR8kZ1WmteI9dvWmagnXRlfsUZW0y0KWe9Ma/ISOa6QPb0Cc7T8LUg2lpREQpRTt119RX9Xugd2uZ+UzbGpqn9u5JY2+vddwv+zFL0g/Vu9h3kd02v+WTXeV8wLDJYkoc6XeTBzC0uQtNJt4oN7hWcMBrkYmnXxslIjGw/nE7eBnG02XWTCeTjADqXtz2j4xMFmzJod+4j4f6MQZBII9Sz5A2FfSbadUuukEYcFvkqLnjG/sMzcBQMIDP6vmy7VRnETahlNi7Pt+gRLHuhGN1BbScolW7a0YnWlI1T0MtKb4DBsjX9GfotxkG+ORg/i45YnG6KPQ6+jgXU3Xtf0tgjTqHVgc7TOCm9R2b4G24RSQXx3WaMP1slbSYsf/YmXIqdg/dVaCFrJdA1v2bDFymuPbA0P4Ey8c4ylInXGgK4yxS7nVC4qbbVtK5bWT3SRaBWw=|0r9TcWfl93tZln8+ZGrPwZTHBbrLiSuVXvoI+cQeSr0=",
    "providerOrganizations": [],
    "providers": [],
    "securityStamp": "6cbf7cd1-d2a2-4c5a-97d7-91042285c49d",
    "twoFactorEnabled": false,
    "usesKeyConnector": false
  },
  "sends": [],
  "unofficialServer": true
}
`))

	httpmock.RegisterResponder("GET", `http://127.0.0.1:8081/api/accounts/profile`,
		httpmock.NewStringResponder(200, `{"_status":0,"avatarColor":null,"culture":"en-US","email":"test-202341-argon2@laverse.net","emailVerified":true,"forcePasswordReset":false,"id":"c64a1f8d-223c-40fd-adde-a90d58dbc26c","key":"2.VBpAJrIHYLv60UTYy5e7sA==|WWsozKnKPVvBYttXtlAzpmZPoffwlY3+8Wup9SBdGcO8T4Ybj808DubDhICTPMz9RliDSJ1OuaivBC0rh/7EcWV1s9KuRth5J3XFWJqmtXU=|l1ly+Uek1j4k2xQNT8iJ++GdwMTWRaSBP5mFvgJ+CGU=","masterPasswordHint":null,"name":"test-202341","object":"profile","organizations":[],"premium":true,"premiumFromOrganization":false,"privateKey":"2.0Xc+uV+6YVa+Zkg74C/MJA==|xTbZGy2r8/kHcXXy23YDHWDTHjasUUoHiWJ84rGEhuPD27GTF8DrFL+lbERo3+OH7MXvseAFygKArvSumlCQnMhT/dswJMH8ZEMuxQUZ8I7kg/7eNqxRndxyC8alZh3VM4FC7TsuVBe6BF4pbPQYX4etFL8yOHLpIhwkKcG7+IfQHQHwOXpsSiGpADQlk5er990snhAXwGpgROqoveO9klpNJXuEzJKMFMdoo9FCaRsJ4bB3BE/Y+8Ph7ek8mGyoUUqNFp8Z/T+XN+9kvxDFnVFtYp1p+sU8LpCSvaq3dAImQt6X7vAgfjbWVkTxB0HlBdMkjpg8BAK7qpscT5oH83SZqdNPRi9kkT9Y30xl0QvvJzXLcjqOS6je+i3vB1r4O8X3ZIj+th4cfZOfHKpzDLbGyAezxiiWwQ4xt0USQ9rLv+BU5YNDKnibmsIooKHuh8wV0qht8vr2CEVavXOmmBi/6bGIWJMOs2K52be5LJkYL+653dsKXBN4uahQDMdfs9vPUA7OoIFR9BZvQiGMDstFYVgJUulOkj2J4nieuTAmurQe6nS6U4v1swI7DIzdo8ZCJcAQfWYWk8IBwh6gQxSuPMg7+O2dntbPFkMP2KhoisUucoXNIDXmxwYAMeP01KMNSgFc6hL6WnLzyVcRlSLy7OC9qwTD6ZH6XA4D/MO9MINTTzw4+/pZvwTeuXMqB0HkcjTpWzTUi405uV6qh8CxYOaNEl6kNAMiuVNASToqYb/EBC0uGOybOzcNZOCiXmuiPVOstiz65KP1d3Dfl9al48hpxnDxFyFub5NDx8xADA3SzOlI9xMCsC/mcY0/fRlwoTDFCLzfdtPIdSn5N8/YySY10e/TXftKGV7bLvzGAlOntwoiWaLJbyHAnEXNUVFxJu5XVaVccXcxIKete42S291dCu4yk35KsIQU0jBaPB/hGXLJvvySl9/kARl0zIXJH+Pk0hlk+/IRH9HNZrru93WTgW9KjEeT2vaZ6UqULkkspIfoUrFxQfSyAxycDaqa4EHt1QJBKyC9+aWEOXNwtxkJTnhtvlqRPOGUpdkAyC5ebfdX2URZnd/2TR/PTviWaDKe/g51UWr8QepgwbJjkBeuZMSPCNHkcLdmwEZERisbZ7H50hDQhhK+qfmMOvkVrRRXKT2ICbQzchm5rDasdtMed7vIf3beS0ESIR8kZ1WmteI9dvWmagnXRlfsUZW0y0KWe9Ma/ISOa6QPb0Cc7T8LUg2lpREQpRTt119RX9Xugd2uZ+UzbGpqn9u5JY2+vddwv+zFL0g/Vu9h3kd02v+WTXeV8wLDJYkoc6XeTBzC0uQtNJt4oN7hWcMBrkYmnXxslIjGw/nE7eBnG02XWTCeTjADqXtz2j4xMFmzJod+4j4f6MQZBII9Sz5A2FfSbadUuukEYcFvkqLnjG/sMzcBQMIDP6vmy7VRnETahlNi7Pt+gRLHuhGN1BbScolW7a0YnWlI1T0MtKb4DBsjX9GfotxkG+ORg/i45YnG6KPQ6+jgXU3Xtf0tgjTqHVgc7TOCm9R2b4G24RSQXx3WaMP1slbSYsf/YmXIqdg/dVaCFrJdA1v2bDFymuPbA0P4Ey8c4ylInXGgK4yxS7nVC4qbbVtK5bWT3SRaBWw=|0r9TcWfl93tZln8+ZGrPwZTHBbrLiSuVXvoI+cQeSr0=","providerOrganizations":[],"providers":[],"securityStamp":"6cbf7cd1-d2a2-4c5a-97d7-91042285c49d","twoFactorEnabled":false,"usesKeyConnector":false}`))

	// We're changing the revisionDate on purpose in the response to mimick Bitwarden official server's behavior
	httpmock.RegisterResponder("POST", `http://127.0.0.1:8081/api/ciphers`,
		httpmock.NewStringResponder(200, `
		{
      "attachments": null,
      "card": null,
      "collectionIds": [],
      "creationDate": "2024-09-22T11:13:40.346903Z",
      "data": {
        "autofillOnPageLoad": null,
        "fields": [
          {
            "linkedId": null,
            "name": "2.svtF3aK9R1MLHt2GwH7Ykw==|PqVDa02T0Vzx6ogTNz1Xyg==|g9/Gd5OOLkVtyVAk9vkRCY2reWTZjdY1D3XsVpBfQA4=",
            "type": 0,
            "value": "2.K49DvcymvaQ/+V/3soCb1Q==|r+b2dz0YqFeA8BrHii5a0DIjkwxdwlY+aEEH2d2NdC0=|KEJAjqImC0Jit9oaFN2rSZHyetxuD7jEjMt07HPFLvU="
          }
        ],
        "name": "2.LjR8NxRtCB1noDILxNKmPQ==|s4e2AO4I3HqmPRRsbsYl0XYSuWu7+sn2K3+jNiFjsW0=|3NQW64/3RmPxe3hhUGeMTrSy9Ruh5hYRlJxIhhWhhSI=",
        "notes": "2.ke3IFCPe50UCM2XdJDS/VQ==|ksO+dyEuRBgrpAelfhxNhw==|JrbMU58V5QyiTpdJXyWB1g2l8jPcqyeWDMqjOnaa9UA=",
        "password": null,
        "passwordHistory": [],
        "passwordRevisionDate": null,
        "totp": "2.mScxVU7uCx3fiPXYlgzWVA==|yxqIjknG71Qdwxq1wyVufWyB1Hb6qWowPgGoZNV2d4s=|uY7hrn7Eow81A76/n4VUrJxSRtC9VDnUdaesO6y0ivY=",
        "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
        "uris": [
          {
            "match": null,
            "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
            "uriChecksum": "2.WfCD+h+rQ49rJDiTM7ycjw==|QgYnnlKzeymjRXmW15SC7bnWWvWaU1iki7rABdi0+RjPgRVGLluT/lQCKXkr5fXH|vaMZaBCKZySaycxVXjpDatKjrN6mBaq2ffRsygLTGzA="
          }
        ],
        "username": "2.9hmypXsnAjVxJ2e5kZQLKA==|8Jif+MQgdTuAlaCn7i/xig==|xy7zJh4qXutESyC02aKSYb/79EmfbGsYlxsYfKqneLA="
      },
      "deletedDate": null,
      "edit": true,
      "favorite": false,
      "fields": [
        {
          "linkedId": null,
          "name": "2.svtF3aK9R1MLHt2GwH7Ykw==|PqVDa02T0Vzx6ogTNz1Xyg==|g9/Gd5OOLkVtyVAk9vkRCY2reWTZjdY1D3XsVpBfQA4=",
          "type": 0,
          "value": "2.K49DvcymvaQ/+V/3soCb1Q==|r+b2dz0YqFeA8BrHii5a0DIjkwxdwlY+aEEH2d2NdC0=|KEJAjqImC0Jit9oaFN2rSZHyetxuD7jEjMt07HPFLvU="
        }
      ],
      "folderId": null,
      "id": "24d1c150-5dfd-4008-964c-01317d1f6b23",
      "identity": null,
      "key": null,
      "login": {
        "autofillOnPageLoad": null,
        "password": null,
        "passwordRevisionDate": null,
        "totp": "2.mScxVU7uCx3fiPXYlgzWVA==|yxqIjknG71Qdwxq1wyVufWyB1Hb6qWowPgGoZNV2d4s=|uY7hrn7Eow81A76/n4VUrJxSRtC9VDnUdaesO6y0ivY=",
        "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
        "uris": [
          {
            "match": null,
            "uri": "2.7ihoxwsJH4HBkTzjFBwFIA==|5EkQHm4IRExHH/LAU6D/jw==|Hs6YgBIFuTQNnH9M3ejgUAVzsGjxAyaVPi1pl1NU9wg=",
            "uriChecksum": "2.WfCD+h+rQ49rJDiTM7ycjw==|QgYnnlKzeymjRXmW15SC7bnWWvWaU1iki7rABdi0+RjPgRVGLluT/lQCKXkr5fXH|vaMZaBCKZySaycxVXjpDatKjrN6mBaq2ffRsygLTGzA="
          }
        ],
        "username": "2.9hmypXsnAjVxJ2e5kZQLKA==|8Jif+MQgdTuAlaCn7i/xig==|xy7zJh4qXutESyC02aKSYb/79EmfbGsYlxsYfKqneLA="
      },
      "name": "2.LjR8NxRtCB1noDILxNKmPQ==|s4e2AO4I3HqmPRRsbsYl0XYSuWu7+sn2K3+jNiFjsW0=|3NQW64/3RmPxe3hhUGeMTrSy9Ruh5hYRlJxIhhWhhSI=",
      "notes": "2.ke3IFCPe50UCM2XdJDS/VQ==|ksO+dyEuRBgrpAelfhxNhw==|JrbMU58V5QyiTpdJXyWB1g2l8jPcqyeWDMqjOnaa9UA=",
      "object": "cipherDetails",
      "organizationId": null,
      "organizationUseTotp": true,
      "passwordHistory": [],
      "reprompt": 0,
      "revisionDate": "2024-09-22T11:13:40.345356Z",
      "secureNote": null,
      "type": 1,
      "viewPassword": true
    }`))

	return webapi.NewClient("http://127.0.0.1:8081/", webapi.WithCustomClient(client), webapi.DisableRetries())
}

func createTestAccount(t *testing.T) {
	ctx := context.Background()
	preloginKey, err := keybuilder.BuildPreloginKey(testPassword, testAccount.Email, testAccount.KdfConfig)
	if err != nil {
		t.Fatal(err)
	}

	hashedPassword := crypto.HashPassword(testPassword, *preloginKey, false)

	block, _ := pem.Decode([]byte(rsaPrivateKey))
	if block == nil {
		t.Fatal(err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	encryptionKeyBytes, err := base64.StdEncoding.DecodeString(encryptionKey)
	if err != nil {
		t.Fatal(err)
	}

	newEncryptionKey, encryptedEncryptionKey, err := keybuilder.EncryptEncryptionKey(*preloginKey, encryptionKeyBytes)
	if err != nil {
		t.Fatal(err)
	}

	publicKey, encryptedPrivateKey, err := keybuilder.EncryptRSAKeyPair(*newEncryptionKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	signupRequest := webapi.SignupRequest{
		Email:              testAccount.Email,
		Name:               testAccount.Email,
		MasterPasswordHash: hashedPassword,
		Key:                encryptedEncryptionKey,
		KdfIterations:      testAccount.KdfConfig.KdfIterations,
		Keys: webapi.KeyPair{
			PublicKey:           publicKey,
			EncryptedPrivateKey: encryptedPrivateKey,
		},
	}

	client := webapi.NewClient("http://127.0.0.1:8080")
	err = client.RegisterUser(ctx, signupRequest)
	if err != nil {
		t.Fatal(err)
	}
}
