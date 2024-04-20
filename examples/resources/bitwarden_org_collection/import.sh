# Find the identifier of the resource you want to import:
#
# $ bw list organizations --search "My Organization" | jq  '.[] .id'
# [{"object":"organization","id":"3b985a2f-0eed-461e-a5ac-adf5015b00c4","...

# $ bw list org-collections --search "Terraform" --organizationid "3b985a2f-0eed-461e-a5ac-adf5015b00c4" | jq  '.[] .id'
# > "94d858f7-03b9-4897-bee1-9af465988932"
#
resource "bitwarden_org_collection" "terraform" {
  name = "Terraform"
}

# Provide those identifiers to Terraform:
$ terraform import bitwarden_org_collection.terraform 3b985a2f-0eed-461e-a5ac-adf5015b00c4/94d858f7-03b9-4897-bee1-9af465988932