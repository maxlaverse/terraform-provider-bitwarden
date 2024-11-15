package embedded

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	mockedServerUrl     = "http://127.0.0.1:8081"
	testDeviceIdentifer = "10a00887-3451-4607-8457-fcbfdc61faaa"
	testDeviceVersion   = "dev"
)

func MockedClient(t testing.TB, name string) webapi.Client {
	t.Helper()

	return webapi.NewClient(mockedServerUrl, testDeviceIdentifer, testDeviceVersion, webapi.WithCustomClient(MockedHTTPClient(t, mockedServerUrl, name)), webapi.DisableRetries())
}

func MockedHTTPClient(t testing.TB, serverUrl string, name string) http.Client {
	t.Helper()

	client := http.Client{Transport: httpmock.DefaultTransport}

	fixturesDir := fixturesDir(t)
	files, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") || !strings.HasPrefix(file.Name(), fmt.Sprintf("%s_", name)) {
			continue
		}

		data, err := os.ReadFile(path.Join(fixturesDir, file.Name()))
		if err != nil {
			t.Fatal(err)
		}

		tmp := strings.Split(file.Name(), "_")
		name = tmp[0]
		method := strings.ToUpper(tmp[1])
		mockUrl := strings.Join(tmp[2:], "/")

		mockUrl, _ = strings.CutSuffix(mockUrl, ".json")
		mockUrl = fmt.Sprintf("%s/%s", serverUrl, mockUrl)

		httpmock.RegisterResponder(method, mockUrl,
			func(req *http.Request) (*http.Response, error) {
				if req.Body != nil {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						return nil, fmt.Errorf("error reading request body '%s %s': %w", method, mockUrl, err)
					}
					if strings.Contains(string(body), "sensitive-") {
						t.Fatalf("Request body contains sensitive information: %s", body)
					}
				}
				return httpmock.NewStringResponse(200, string(data)), nil
			},
		)
	}

	return client
}

func fixturesDir(t testing.TB) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get caller information")
	}
	dir := filepath.Dir(file)
	dir = path.Join(dir, "fixtures")
	return dir
}
