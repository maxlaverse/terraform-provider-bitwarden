# Find the identifier of the resource you want to import:
#
# $ bw list items --search "Mysql Root Credentials" | jq  '.[] .id'
# ? Master password: [hidden]
#
# > "ec4e447f-9aed-4203-b834-c8f3848828f7"
#

# Provide this identifier to Terraform:
$ terraform import bitwarden_item_login.administrative-user ec4e447f-9aed-4203-b834-c8f3848828f7