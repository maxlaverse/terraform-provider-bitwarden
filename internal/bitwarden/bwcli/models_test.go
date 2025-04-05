//go:build offline

package bwcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVaultFromServer(t *testing.T) {
	testCases := []struct {
		vaultUrl       string
		providerUrl    string
		expectedResult bool
	}{
		{
			vaultUrl:       "http://127.0.0.1",
			providerUrl:    "http://127.0.0.1",
			expectedResult: true,
		},
		{
			vaultUrl:       "http://127.0.0.1/",
			providerUrl:    "http://127.0.0.1",
			expectedResult: true,
		},
		{
			vaultUrl:       "http://127.0.0.1",
			providerUrl:    "http://127.0.0.1/",
			expectedResult: true,
		},
		{
			vaultUrl:       "",
			providerUrl:    "https://vault.bitwarden.com",
			expectedResult: true,
		},
		{
			vaultUrl:       "https://vault.bitwarden.com/",
			providerUrl:    "https://vault.bitwarden.com",
			expectedResult: true,
		},
		{
			vaultUrl:       "http://127.0.0.1",
			providerUrl:    "http://127.0.0.1/s",
			expectedResult: false,
		},
		{
			vaultUrl:       "http://127.0.0.2",
			providerUrl:    "http://127.0.0.1",
			expectedResult: false,
		},
	}

	for _, test := range testCases {
		t.Run("", func(t *testing.T) {
			status := &Status{
				ServerURL: test.vaultUrl,
			}
			match := status.VaultFromServer(test.providerUrl)

			assert.Equal(t, test.expectedResult, match)
		})
	}

}
