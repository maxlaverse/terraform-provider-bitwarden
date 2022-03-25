package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func hasOneInstanceState(s []*terraform.InstanceState) error {
	if len(s) != 1 {
		return fmt.Errorf("expected %d states, got %d: %+v", 1, len(s), s)
	}
	return nil
}
