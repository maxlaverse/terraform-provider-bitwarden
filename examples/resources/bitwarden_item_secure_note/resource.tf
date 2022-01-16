resource "bitwarden_item_secure_note" "exampleservice-configuration" {
  name  = "ExampleService Configuration"
  notes = <<EOT
[global]
secret = "<sensitive>"
EOT
}
