package embedded

import (
	"strings"

	"github.com/hashicorp/go-version"
)

// cipherKeyEncMinServerVer is CIPHER_KEY_ENC_MIN_SERVER_VER in bitwarden/clients.
const cipherKeyEncMinServerVer = "2024.2.0"

var cipherKeyEncMin = version.Must(version.NewVersion(cipherKeyEncMinServerVer))

// cipherKeyEncryptionEnabled reports whether new items should get a per-cipher key.
// Official clients gate on server version after removing the cipher-key-encryption
// feature flag; the flag is kept as a fallback for missing or unparseable versions.
func cipherKeyEncryptionEnabled(serverVersion string, featureFlag bool) bool {
	if parsed, err := version.NewVersion(strings.TrimSpace(serverVersion)); err == nil && parsed.GreaterThanOrEqual(cipherKeyEncMin) {
		return true
	}
	return featureFlag
}
