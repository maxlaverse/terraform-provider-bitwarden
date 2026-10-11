package webapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

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

// /api uses {"object":"error","message":"..."}. /identity/connect/token uses RFC 6749
// §5.2 ({"error","error_description"}), sometimes with ErrorModel.
type httpErrorBody struct {
	Object           string `json:"object"`
	Message          string `json:"message"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorModel       struct {
		Message string `json:"message"`
	} `json:"ErrorModel"`
}

func httpErrorFromResponse(httpReq *http.Request, httpResp *http.Response, respBody []byte) *HTTPError {
	if msg, ok := recognizedHTTPError(httpResp.StatusCode, respBody); ok {
		return &HTTPError{StatusCode: httpResp.StatusCode, Message: msg}
	}
	return &HTTPError{
		StatusCode: httpResp.StatusCode,
		Message:    fmt.Sprintf("bad response status code for '%s %s': %d!=200", httpReq.Method, httpReq.URL, httpResp.StatusCode),
	}
}

func recognizedHTTPError(status int, body []byte) (string, bool) {
	var parsed httpErrorBody
	if json.Unmarshal(body, &parsed) != nil {
		return "", false
	}
	if parsed.Object == "error" && parsed.Message != "" {
		return fmt.Sprintf("the server returned an error: \"%s\" (%d)", parsed.Message, status), true
	}
	if !isOAuthTokenError(parsed.Error) {
		return "", false
	}
	msg := parsed.ErrorDescription
	if msg == "" {
		msg = parsed.ErrorModel.Message
	}
	if msg == "" {
		msg = parsed.Error
	}
	return msg, true
}

func isOAuthTokenError(code string) bool {
	switch code {
	case "invalid_request", "invalid_client", "invalid_grant",
		"unauthorized_client", "unsupported_grant_type", "invalid_scope":
		return true
	default:
		return false
	}
}
