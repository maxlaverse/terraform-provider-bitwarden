data "bitwarden_secret" "example" {
  id = "37a66d6a-96c1-4f04-9a3c-b1fc0135669e"
}

resource "kubernetes_secret" "vpn_credentials" {
  metadata {
    name = "vpn-key"
  }

  data = {
    "PASSWORD" = data.bitwarden_secret.value
  }
}
