# Find the identifier of the resource you want to import:
#
# $ bw list items --search "SSH Private Key" | jq  '.[] .id'
# ? Master password: [hidden]
#
# > "a9e19f26-1b8c-4568-bc09-191e2cf56ed6"
#

# Provide this identifier to Terraform:
$ terraform import bitwarden_item_secure_note.ssh-private-key a9e19f26-1b8c-4568-bc09-191e2cf56ed6