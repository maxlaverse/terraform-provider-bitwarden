data "bitwarden_project" "example" {
  id = "37a66d6a-96c1-4f04-9a3c-b1fc0135669e"
}

resource "bitwarden_secret" "example" {
  project_id = data.bitwarden_project.example.id

  key   = "ACCESS_KEY"
  value = "THIS-VALUE"
}
