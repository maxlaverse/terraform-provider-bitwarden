# Find the identifier of the resource you want to read from, with the Bitwarden CLI:
#
# $ bw list items --search "Mysql Root Credentials" | jq
# [
#   {
#     "object": "item",
#     "id": "ec4e447f-9aed-4203-b834-c8f3848828f7",
#     "organizationId": null,
#     "folderId": null,
#     "type": 1,
#     "reprompt": 0,
#     "name": "Mysql Root Credentials",
#     "notes": null,
#     "favorite": false,
#     "login": {
#       "username": "root",
#       "password": "<sensitive>",
#       "totp": null,
#       "passwordRevisionDate": null
#     },
#     "collectionIds": [],
#     "revisionDate": "2020-10-23T18:09:31.426Z"
#   }
# ]

data "bitwarden_item_login" "mysql-root-credentials" {
  id = "ec4e447f-9aed-4203-b834-c8f3848828f7"
}

# Later to be accessed as
#   ... = data.bitwarden_item_login.mysql-root-credentials.username
#   ... = data.bitwarden_item_login.mysql-root-credentials.password
