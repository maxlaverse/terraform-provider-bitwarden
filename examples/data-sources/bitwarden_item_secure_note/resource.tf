# Find the identifier of the resource you want to read from, with the Bitwarden CLI:
#
# $ bw list items --search "SSH Private Key" | jq
# [
#   {
#     "object": "item",
#     "id": "a9e19f26-1b8c-4568-bc09-191e2cf56ed6",
#     "organizationId": null,
#     "folderId": null,
#     "type": 2,
#     "reprompt": 0,
#     "name": "SSH Private Key",
#     "notes": "<sensitive>",
#     "favorite": false,
#     "secureNote": {
#       "type": 0
#     },
#     "collectionIds": [],
#     "revisionDate": "2021-09-16T20:56:42.946Z"
#   }
# ]

data "bitwarden_item_secure_note" "ssh-private-key" {
  id = "a9e19f26-1b8c-4568-bc09-191e2cf56ed6"
}

# Later to be accessed as
#   ... = data.bitwarden_item_secure_note.ssh-private-key.notes
