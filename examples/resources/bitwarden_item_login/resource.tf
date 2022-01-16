resource "bitwarden_item_login" "administrative-user" {
  name     = "Service Administrator"
  username = "admin"
  password = "<sensitive>"
  totp     = "<sensitive>"
  notes    = "some notes about this user"
  folder   = "3b985a2f-0eed-461e-a5ac-adf5015b00c4"
}
