# Bitwarden Master Password
variable "bw_password" {
  type        = string
  description = "Bitwarden Master Key"
  sensitive   = true
}

# Bitwarden API credentials when using the official bitwarden.com Vault
#
# 1. Connect to your account on https://vault.bitwarden.com
# 2. Click on "Settings" and then "My Account"
# 3. Scroll down to the "API Key" section
# 4. Click on "View API Key" (or maybe another label if it's the first time)
# 5. Save the API credentials
# 6. Before running `terraform apply`, export the API credentials as environment variable:
#    export TF_VAR_bw_client_id=<client_id>
#    export TF_VAR_bw_client_secret=<client_secret>
#    export TF_VAR_bw_password=<master_password>
#
variable "bw_client_id" {
  type        = string
  description = "Bitwarden Client ID"
  sensitive   = true
}

variable "bw_client_secret" {
  type        = string
  description = "Bitwarden Client Secret"
  sensitive   = true
}