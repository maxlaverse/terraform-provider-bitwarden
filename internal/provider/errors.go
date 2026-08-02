package provider

import "errors"

var (
	errPasswordManagerRequired = errors.New("provider was not configured with Password Manager credentials")
	errSecretsManagerRequired  = errors.New("provider was not configured with Secrets Manager credentials")
)
