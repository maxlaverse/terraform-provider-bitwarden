package webapi

import "net/http"

type Options func(c Client)

func DisableRetries() Options {
	return func(c Client) {
		c.(*client).httpClient.RetryMax = 0
	}
}

func WithCustomClient(httpClient http.Client) Options {
	return func(c Client) {
		c.(*client).httpClient.HTTPClient = &httpClient
	}
}

func WithDeviceIdentifier(deviceIdentifier string) Options {
	return func(c Client) {
		c.(*client).deviceIdentifier = deviceIdentifier
	}
}
