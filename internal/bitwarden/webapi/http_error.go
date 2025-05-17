package webapi

import "errors"

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func (e *HTTPError) GetStatusCode() int {
	return e.StatusCode
}

func IsHTTPError(err error) (*HTTPError, bool) {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr, true
	}
	return nil, false
}
