// Package provider implements the Bitwarden Terraform provider.
//
// Migration note: the provider is temporarily served via terraform-plugin-mux,
// combining a Plugin Framework shell (New) with the SDKv2 implementation (NewSDK).
// Move resources/data sources to Framework incrementally; remove NewSDK and mux
// when nothing remains on SDKv2.
package provider
