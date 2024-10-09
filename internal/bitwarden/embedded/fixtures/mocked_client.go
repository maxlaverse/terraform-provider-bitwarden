package fixtures

import (
	"fmt"
	"net/http"
	"os"
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

func MockedClient(t *testing.T, name string) webapi.Client {
	return webapi.NewClient(mockedServerUrl, testDeviceIdentifer, testDeviceVersion, webapi.WithCustomClient(MockedHTTPClient(t, mockedServerUrl, name)), webapi.DisableRetries())
}

func MockedHTTPClient(t *testing.T, serverUrl string, name string) http.Client {
	client := http.Client{Transport: httpmock.DefaultTransport}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get caller information")
	}
	dir := filepath.Dir(file)

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") || !strings.HasPrefix(file.Name(), fmt.Sprintf("%s_", name)) {
			continue
		}

		data, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, file.Name()))
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
			httpmock.NewStringResponder(200, string(data)))
	}

	return client
}
