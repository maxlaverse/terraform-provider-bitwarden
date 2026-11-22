//go:build offline

package embedded

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCipherKeyEncryptionEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		flag    bool
		want    bool
	}{
		{name: "empty version flag false", version: "", flag: false, want: false},
		{name: "2023.12.0 flag false", version: "2023.12.0", flag: false, want: false},
		{name: "2023.12.0 flag true", version: "2023.12.0", flag: true, want: true},
		{name: "2024.2.0 flag false", version: "2024.2.0", flag: false, want: true},
		{name: "2024.1.9 flag false", version: "2024.1.9", flag: false, want: false},
		{name: "2025.6.1 flag false", version: "2025.6.1", flag: false, want: true},
		{name: "garbage version flag false", version: "not-a-version", flag: false, want: false},
		{name: "garbage version flag true", version: "not-a-version", flag: true, want: true},
		{name: "prerelease suffix on new version", version: "2025.6.1-hotfix", flag: false, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, cipherKeyEncryptionEnabled(tt.version, tt.flag))
		})
	}
}
