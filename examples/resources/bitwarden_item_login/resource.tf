resource "bitwarden_item_login" "administrative-user" {
  name            = "Service Administrator"
  username        = "admin"
  password        = "<sensitive>"
  totp            = "<sensitive>"
  notes           = "some notes about this user"
  folder_id       = "3b985a2f-0eed-461e-a5ac-adf5015b00c4"
  favorite        = true

  field {
    name = "this-is-a-text-field"
    text = "text-value"
  }

  field {
    name    = "this-is-a-boolean-field"
    boolean = true
  }

  field {
    name   = "this-is-a-hidden-field"
    hidden = "text-value"
  }
}
