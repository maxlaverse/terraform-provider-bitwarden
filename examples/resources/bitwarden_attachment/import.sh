# Find the identifier of the resource and its attachment you want to import:
#
# $ bw list items --search "SSH Credentials" | jq  '.[]'
# ? Master password: [hidden]
#
# > {
# >   "object": "item",
# >   "id": "59575167-4d36-5a58-466e-d9021926df8a",
# >   [...]
# >   "name": "My Top Secret SSH Credentials",
# >   "attachments": [
# >     {
# >       id": "4d6a41364d6a4dea8ddb1a",
# >       "fileName": "ssh_private.key",
# >       "size": "1743",
# >       "sizeName": "1.74 KB",
# >       "url": "https://vault.bitwarden.com/attachments/59575167-4d36-5a58-466e-d9021926df8a/4d6a41364d6a4dea8ddb1a"
# >     }
# >   ],
# > }

resource "bitwarden_attachment" "ssh_private" {
  # ...instance configuration...
}

# Provide both identifiers to Terraform in the form of '<item_id>/<attachment_id>'
$ terraform import bitwarden_attachment.ssh_private 59575167-4d36-5a58-466e-d9021926df8a/4d6a41364d6a4dea8ddb1a