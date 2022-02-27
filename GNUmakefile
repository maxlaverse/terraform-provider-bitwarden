default: testacc

# Run acceptance tests
.PHONY: testacc docs tffmt
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

tffmt:
	terraform fmt  -recursive examples

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
	find docs -type f -exec sed -i '' '/INTERNAL USE/d' {} \;