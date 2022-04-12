# Find the identifier of the resource you want to import:
#
# $ bw list folders --search "us-west-2" | jq  '.[] .id'
# ? Master password: [hidden]
#
# > "94d858f7-03b9-4897-bee1-9af465988932"
#

# Provide this identifier to Terraform:
$ terraform import bitwarden_folder.us-west-2 94d858f7-03b9-4897-bee1-9af465988932