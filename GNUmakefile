default: testacc

# Run acceptance tests
.PHONY: testacc docs tffmt
testacc: clean
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

tffmt:
	terraform fmt  -recursive examples

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.19.0
	find docs -type f -name '*.md' -exec sed -i '' '/INTERNAL USE/d' {} \;

clean:
	rm internal/provider/.bitwarden/data.json || true

server:
	docker run -ti \
	-e I_REALLY_WANT_VOLATILE_STORAGE=true \
	-e DISABLE_ICON_DOWNLOAD=true \
	-e ADMIN_TOKEN=test1234 \
	-e LOGIN_RATELIMIT_SECONDS=1 \
	-e LOGIN_RATELIMIT_MAX_BURST=1000000 \
	-e ADMIN_RATELIMIT_SECONDS=1 \
	-e ADMIN_RATELIMIT_MAX_BURST=1000000 \
	--mount type=tmpfs,destination=/data \
	-p 8080:80 vaultwarden/server:1.32.5
