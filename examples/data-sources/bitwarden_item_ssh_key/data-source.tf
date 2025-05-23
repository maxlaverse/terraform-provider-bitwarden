data "bitwarden_item_ssh_key" "git_ssh_key" {
  search = "Git/SSH Key"
}

resource "local_sensitive_file" "id_rsa" {
  filename        = "id_rsa"
  file_permission = "600"
  content         = data.bitwarden_item_ssh_key.id_rsa.private_key
}