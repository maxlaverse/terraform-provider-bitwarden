package test

type KeyPair struct {
	PublicKey           string `json:"publicKey"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
}

type SignupRequest struct {
	Email              string  `json:"email"`
	Name               string  `json:"name"`
	MasterPasswordHash string  `json:"masterPasswordHash"`
	Key                string  `json:"key"`
	Kdf                int     `json:"kdf"`
	KdfIterations      int     `json:"kdfIterations"`
	Keys               KeyPair `json:"keys"`
}
