resource "bitwarden_item_secure_note" "my-service-configuration" {
  name  = "Service Administrator"
  notes = <<EOT
[global]
secret = "<sensitive>"
EOT
}
