package test

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
)

/*
* This is a client to interact with the Vaultwarden or eventually
* Bitwarden compatible API. It's only meant to be used for test
* purposes and is to be considered as an insecure implementation
 */

const (
	testPrivatePEM = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAtY3IxI7rgD/GubkWzfB5kr0JRQUD+V89JSy9fSXK4hX+vAv3
j7ifsmB7Ti3z52+0Op8h3VmskjWKuj49YWneSS8fd5yWv+u+BhthcR5mzQhFokYr
1ofpw+ZDb4QzpNq4GJdYIlGVZWamszsyP1bMfZj6SzuJWUWkF2fLvhYc/sv2xMa3
yPOxlOxSmBTSFCOWeenbneMIiYSkMGPuzInc/4MP0PgkX3oXLqXN+mPAH8BVEKo1
n7dcjTCqweD2rNJNTmbPDR8svbRnPEj6ZakqR0W7DFR4nL1ABM6CpBysJDrJRRjo
2BU/UFDt3gxZBixySZyGPOfgjR1hHW9IA80uawIDAQABAoIBADF8JtfkeK4pM/pV
R7D0Nb7YRZmJZ27mFZ13V2KwFV+QTTFmNaD32pddEt7ZSZywZrh/vVQ+5/mmHFzt
L//IQm1CKdqJVNGU6ONzPRj/B1glRA2aAMLlosLhlUnlB8qMTxds0PnxsL3Fv/Qk
U2rONsiZFMfCv4oUoEg841y5XEfaTKcXaY1MQDWn1egaSLGs+S+LKl8d6jD15gnA
wlPobK7rbxYK53+rwRS7VlpPun5siqjjtnUCoT0HPZ1qb0I08PZ/PIcvldH18GMz
Gido3znpAmDY0fQTxicp1lR5MyV8GvhBBJ7gFVq7m0njCyni9h9YiJq8Lhtjr0Af
wfDSDukCgYEA3LS/vM8WL2kkzcVdOnLoM0qQ9kJ8BcRkM6qPexuddAUKk7gh6tOS
u5cLCp2AXQNSCMHlnqJUjUeqTvdbzWgHBwVkFvR9LqpeoCld+i6MrlbsZt1/PfcU
5Cq+UTBfZ/Pe5x/fF1tmhvwTvJafg28WrVmZmFw/5hrD4ti5163FYBUCgYEA0pY5
gNlOsJdoIs/vJLGAtb165BeCy3ei1w3LOK35ixxRykghV3nxaFcQkLFBY5FF/0yn
6GyYehAE7GJuZLhYLW2uerk6SoZykHDlRw/X16mMh8rBLrXpys49dWvHIboFkuNf
q0gE62Tkfus90AfXh0BQjFlmg+PXlc2Z6E81dH8CgYEAqI0ZSRZV+QsxYjxyAGs0
zccKgicwFC9x3stJHFlwm+Qlub6LmIzPqJenhQnXuDEK+0kpFUcfj23FsNzTrUDe
7Qu+7pD08SiHb4VoEeJu6c3UaJKL1ETYHZBPHC33Dqp99sCuXWYeHMRyRjo5w+SY
yvZ8iJEa855JLvsYopBBBikCgYBtmcwR2IfQ9uw2+hvP8CY58IUOQ4JKXVi+Lqqv
NDTlhva2nfXkbk4LbQztEaQjqw9QQVg+ao6tMLsvQEeOWjdiZWxi6RaChRkJPgjG
hGNlFhRS9F647erhJ5frDg4U6plOCtLW9WPCE7+sosiIBhzRgtKpSTpGuIWSrPBG
bTs4BwKBgQDCNK/pjXVjyZZX+KJxp81VzB6mwpsHxzr9iuyybU6i5gb+iFBmynhO
ZR0i8ARbNzVgierwIbExBNUtMfcfNZx+HLGFdV5iA/yP8sNZBoexVXwThhaIpQJq
sHn6jLmPgcGLu7FckdRXHv+qzIojOnPXeLhQ3nk+9Cu09NVzL2XBdg==
-----END RSA PRIVATE KEY-----
`
)

type Client interface {
	RegisterUser(name, username, password string, kdfIterations int) error
}

func NewVaultwardenTestClient(serverURL string) Client {
	block, _ := pem.Decode([]byte(testPrivatePEM))
	if block == nil {
		panic("bug - bad private key constant for tests")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	if priv.Size()*8 != 2048 {
		panic(fmt.Sprintf("bug - bad private key size: %d", priv.Size()*8))
	}
	return &client{
		crypto: crypto{
			randomEncryptionKey: []byte{40, 201, 91, 4, 62, 57, 230, 98, 146, 113, 111, 129, 180, 230, 116, 91, 110, 163, 34, 47, 127, 131, 59, 252, 7, 101, 153, 48, 185, 209, 19, 45, 227, 232, 133, 165, 156, 157, 9, 202, 36, 235, 96, 151, 31, 27, 38, 238, 213, 219, 189, 229, 182, 208, 39, 208, 53, 69, 204, 22, 157, 76, 151, 209},
			randomIV:            "A/89QpHf5lcmfJhoEHKbLw==",
		},
		privateKey: priv,
		serverURL:  serverURL,
	}
}

type client struct {
	crypto     crypto
	serverURL  string
	privateKey *rsa.PrivateKey
}

type Obj struct {
	IV   []byte
	Data []byte
	Mac  []byte
	Key  SymmetricCryptoKey
}

func (c *client) RegisterUser(name, username, password string, kdfIterations int) error {
	preloginKey, err := c.crypto.MakePreloginKey(password, username, kdfIterations)
	if err != nil {
		return fmt.Errorf("unable to make prelogin key: %w", err)
	}

	hashedPassword := hashPassword(password, *preloginKey, false)
	encryptionKey, encString, err := c.crypto.MakeEncryptionKey(*preloginKey)
	if err != nil {
		return fmt.Errorf("unable to make encryption key: %w", err)
	}

	keyPair, err := c.crypto.MakeKeyPair(*encryptionKey, c.privateKey)
	if err != nil {
		return fmt.Errorf("unable to make key pair: %w", err)
	}

	reqBody := SignupRequest{
		Email:              username,
		Name:               name,
		MasterPasswordHash: hashedPassword,
		Key:                string(encString),
		KdfIterations:      kdfIterations,
		Keys:               *keyPair,
	}

	testSignupRequest, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("unable to marshall user registration request: %w", err)
	}

	signupUrl := fmt.Sprintf("%s/api/accounts/register", c.serverURL)

	resp, err := http.Post(signupUrl, "application/json", bytes.NewBuffer(testSignupRequest))
	if err != nil {
		return fmt.Errorf("error during registration call: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading registration call response body: %w", err)
	}
	strBody := string(body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code for registration call: %d, %s", resp.StatusCode, strBody)
	}
	return nil
}
