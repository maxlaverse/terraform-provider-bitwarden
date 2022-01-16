resource "bitwarden_item_login" "my-administrative-user" {
  name     = "Service Administrator"
  username = "admin"
  password = "<sensitive>"
  totp     = "<sensitive>"
  notes    = "some notes about this item"
  folder   = "3b985a2f-0eed-461e-a5ac-adf5015b00c4"
}
