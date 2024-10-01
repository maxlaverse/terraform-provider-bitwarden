package webapi

import (
	"context"
	"net"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func CustomRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	debugInfo := map[string]interface{}{
		"error": err,
	}
	if resp != nil {
		debugInfo["status_code"] = resp.StatusCode
		debugInfo["status_message"] = resp.Status
	}

	var willRetry bool
	var handlerErr error

	if err != nil {
		if err == context.DeadlineExceeded {
			willRetry = false
		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			willRetry = false
		}
	} else {
		willRetry, handlerErr = retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	debugInfo["will_retry"] = willRetry
	debugInfo["handler_error"] = handlerErr
	tflog.Trace(ctx, "retry_handler", debugInfo)

	return willRetry, handlerErr
}
