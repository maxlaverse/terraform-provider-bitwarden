// Package provider implements the Bitwarden Terraform provider.
//
// Migration note: the provider is temporarily served via terraform-plugin-mux,
// combining a Plugin Framework shell (New) with the SDKv2 implementation (NewSDK).
// Shared AttrData transforms plus MapData bridge Framework models to that mapping
// layer. Resources/data sources move to Framework incrementally (folder, project,
// then org data sources); remove NewSDK and mux when nothing remains on SDKv2.
package provider
