resource "tls_private_key" "id_rsa" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "bitwarden_item_ssh_key" "id_rsa" {
  name = "Git/SSH Key"

  private_key     = tls_private_key.id_rsa.private_key_pem
  public_key      = tls_private_key.id_rsa.public_key_openssh
  key_fingerprint = tls_private_key.id_rsa.public_key_fingerprint_md5

}