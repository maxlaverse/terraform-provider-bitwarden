default: testacc

# Run acceptance tests
.PHONY: testacc docs tffmt
testacc: clean
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

tffmt:
	terraform fmt  -recursive examples

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.19.0

clean:
	rm internal/provider/.bitwarden/data.json || true

server:
	docker-compose up
