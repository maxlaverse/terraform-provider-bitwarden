# Find the identifier of the resource you want to import:
#
# $ bw list items --search "Mysql Root Credentials" | jq  '.[] .id'
# ? Master password: [hidden]
#
# > "ec4e447f-9aed-4203-b834-c8f3848828f7"
#

data "bitwarden_item_login" "database_credentials" {
  id = "ec4e447f-9aed-4203-b834-c8f3848828f7"
}

# Later to be accessed as
#   ... = data.bitwarden_item_login.database_credentials.username
#   ... = data.bitwarden_item_login.database_credentials.password
#
# or for fields:
# lookup(
#    zipmap(
#      data.bitwarden_item_login.database_credentials.field.*.name,
#      data.bitwarden_item_login.database_credentials.field.*
#    ),
#    "<name-of-the-field-your-looking-for>"
# )