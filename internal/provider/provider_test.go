package provider

import (
	"testing"
)

func TestProviderValidity(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
