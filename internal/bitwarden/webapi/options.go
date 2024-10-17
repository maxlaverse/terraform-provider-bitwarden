package webapi

import "net/http"

type Options func(c Client)

func DisableRetries() Options {
	return func(c Client) {
		roundTripper, ok := c.(*client).httpClient.Transport.(*RetryRoundTripper)
		if !ok {
			return
		}
		roundTripper.DisableRetries = true
	}
}

func WithCustomClient(httpClient http.Client) Options {
	return func(c Client) {
		c.(*client).httpClient = &httpClient
	}
}
