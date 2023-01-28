# Find the identifier of the resource you want to import:
#
# $ bw list items --search "SSH Private Key" | jq  '.[] .id'
# ? Master password: [hidden]
#
# > "a9e19f26-1b8c-4568-bc09-191e2cf56ed6"
#

data "bitwarden_item_secure_note" "ssh_notes" {
  id = "a9e19f26-1b8c-4568-bc09-191e2cf56ed6"
}

# Later to be accessed as
#   ... = data.bitwarden_item_secure_note.ssh_notes.notes
#
# or for fields:
# lookup(
#    zipmap(
#      data.bitwarden_item_secure_note.ssh_notes.field.*.name,
#      data.bitwarden_item_secure_note.ssh_notes.field.*
#    ),
#    "<name-of-the-field-your-looking-for>"
# )