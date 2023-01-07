# Find the identifier of the resource you want to import:
#
# $ bw list folders --search "My Cloud Credentials" | jq  '.[] .id'
# ? Master password: [hidden]
#
# > "94d858f7-03b9-4897-bee1-9af465988932"
#
resource "bitwarden_folder" "cloud_credentials" {
  name = "My Cloud Credentials"
}

# Provide this identifier to Terraform:
$ terraform import bitwarden_folder.cloud_credentials 94d858f7-03b9-4897-bee1-9af465988932