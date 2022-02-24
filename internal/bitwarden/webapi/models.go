package webapi

type SignupRequest struct {
	Email              string  `json:"email"`
	Name               string  `json:"name"`
	MasterPasswordHash string  `json:"masterPasswordHash"`
	Key                string  `json:"key"`
	Kdf                int     `json:"kdf"`
	KdfIterations      int     `json:"kdfIterations"`
	Keys               KeyPair `json:"keys"`
}

type KeyPair struct {
	PublicKey           string `json:"publicKey"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
}

type CreateOrganizationRequest struct {
	Name           string  `json:"name"`
	BillingEmail   string  `json:"billingEmail"`
	PlanType       int     `json:"planType"`
	CollectionName string  `json:"collectionName"`
	Key            string  `json:"key"`
	Keys           KeyPair `json:"keys"`
}

type CreateOrganizationResponse struct {
	Id string `json:"id"`
}

type TokenResponse struct {
	Kdf                 int    `json:"Kdf"`
	KdfIterations       int    `json:"KdfIterations"`
	Key                 string `json:"Key"`
	PrivateKey          string `json:"PrivateKey"`
	ResetMasterPassword bool   `json:"ResetMasterPassword"`
	AccessToken         string `json:"access_token"`
	ExpireIn            int    `json:"expires_in"`
	RefreshToken        string `json:"refresh_token"`
	Scope               string `json:"scope"`
	TokenType           string `json:"token_type"`
	UnofficialServer    bool   `json:"unofficialServer"`
}

type CollectionResponse struct {
	Data   []Collection `json:"data"`
	Object string       `json:"object"`
}

type Collection struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	OrganizationId int    `json:"organization_id"`
	Object         string `json:"object"`
	ExternalId     string `json:"external_id"`
}
