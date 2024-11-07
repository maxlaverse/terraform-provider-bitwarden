package embedded

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

type Account struct {
	AccountUUID            string                  `json:"accountUuid,omitempty"`
	Email                  string                  `json:"email,omitempty"`
	VaultFormat            string                  `json:"vaultFormat,omitempty"`
	KdfConfig              models.KdfConfiguration `json:"kdfConfig,omitempty"`
	ProtectedSymmetricKey  string                  `json:"protectedSymmetricKey,omitempty"`
	ProtectedRSAPrivateKey string                  `json:"protectedRSAPrivateKey,omitempty"`
	Secrets                AccountSecrets          `json:"-"`
}

func (a *Account) LoggedIn() bool {
	return len(a.ProtectedRSAPrivateKey) > 0
}

func (a *Account) SecretsLoaded() bool {
	return len(a.Secrets.MainKey.Key) > 0
}

func (a *Account) ToJSON() string {
	out, _ := json.Marshal(a)
	return string(out)
}

type AccountSecrets struct {
	MasterPasswordHash  string
	MainKey             symmetrickey.Key
	OrganizationSecrets map[string]OrganizationSecret
	RSAPrivateKey       *rsa.PrivateKey
}

func (s *AccountSecrets) GetOrganizationKey(orgId string) (*symmetrickey.Key, error) {
	orgSecret, ok := s.OrganizationSecrets[orgId]
	if !ok {
		return nil, fmt.Errorf("no organization key found for '%s'", orgId)
	}
	return &orgSecret.Key, nil
}

type OrganizationSecret struct {
	Key              symmetrickey.Key
	OrganizationUUID string
	Name             string
}

type MachineAccountClaims struct {
	Organization string `json:"organization"`
	jwt.RegisteredClaims
}
