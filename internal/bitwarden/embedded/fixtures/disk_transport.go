package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type diskTransport struct {
	Transport http.RoundTripper
	Prefix    string
}

func (d *diskTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := d.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if err := d.saveResponseToFile(req, resp); err != nil {
		return nil, fmt.Errorf("error saving response to file: %w", err)
	}

	return resp, nil
}

func (d *diskTransport) saveResponseToFile(req *http.Request, resp *http.Response) error {
	filename := fmt.Sprintf("%s_%s%s.json", d.Prefix, req.Method, sanitizeFilename(resp.Request.URL.EscapedPath()))

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body = io.NopCloser(bytes.NewReader(data))

	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return err
	}

	prettyData, err := json.MarshalIndent(jsonData, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, prettyData, 0644)
}

func sanitizeFilename(url string) string {
	replacer := []string{
		":", "_",
		"/", "_",
		"\\", "_",
		"?", "_",
		"&", "_",
		"=", "_",
		"%", "_",
	}
	replacerMap := strings.NewReplacer(replacer...)
	return replacerMap.Replace(url)
}
