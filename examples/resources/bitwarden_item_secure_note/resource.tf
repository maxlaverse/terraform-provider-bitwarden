resource "bitwarden_item_secure_note" "exampleservice-configuration" {
  name            = "ExampleService Configuration"
  notes           = <<EOT
[global]
secret = "<sensitive>"
EOT
  folder_id       = "3b985a2f-0eed-461e-a5ac-adf5015b00c4"
  organization_id = "54421e78-95cb-40c4-a257-17231a7b6207"
  favorite        = true
  collection_ids  = ["c74d6067-50b0-4427-bec8-483f3270fde3"]

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
