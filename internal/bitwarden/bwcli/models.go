package bwcli

import (
	"strings"
	"time"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden"
)

type VaultStatus string

const (
	StatusLocked          VaultStatus = "locked"
	StatusUnauthenticated VaultStatus = "unauthenticated"
	StatusUnlocked        VaultStatus = "unlocked"
)

type Status struct {
	ServerURL string      `json:"serverURL,omitempty"`
	LastSync  time.Time   `json:"lastSync,omitempty"`
	UserEmail string      `json:"userEmail,omitempty"`
	UserID    string      `json:"userID,omitempty"`
	Status    VaultStatus `json:"status,omitempty"`
}

func (s *Status) VaultFromServer(serverUrl string) bool {
	providerServerUrl := trimSlashSuffix(serverUrl)
	vaultServerUrl := trimSlashSuffix(s.ServerURL)
	return vaultServerUrl == providerServerUrl || len(vaultServerUrl) == 0 && providerServerUrl == bitwarden.DefaultBitwardenServerURL
}

func trimSlashSuffix(serverUrl string) string {
	return strings.TrimSuffix(serverUrl, "/")
}
