package fixtures

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	mockedServerUrl     = "http://127.0.0.1:8081"
	testDeviceIdentifer = "10a00887-3451-4607-8457-fcbfdc61faaa"
)

func MockedClient(t *testing.T, name string) webapi.Client {
	client := http.Client{Transport: httpmock.DefaultTransport}

	files, err := os.ReadDir("fixtures")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Found %d responders", len(files))
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") || !strings.HasPrefix(file.Name(), fmt.Sprintf("%s_", name)) {
			continue
		}

		data, err := os.ReadFile(fmt.Sprintf("%s/%s", "fixtures", file.Name()))
		if err != nil {
			t.Fatal(err)
		}

		tmp := strings.Split(file.Name(), "_")
		name = tmp[0]
		method := strings.ToUpper(tmp[1])
		mockUrl := strings.Join(tmp[2:], "/")

		mockUrl, _ = strings.CutSuffix(mockUrl, ".json")
		mockUrl = fmt.Sprintf("%s/%s", mockedServerUrl, mockUrl)
		t.Logf("Registering responder for %s %s", method, mockUrl)

		httpmock.RegisterResponder(method, mockUrl,
			httpmock.NewStringResponder(200, string(data)))
	}

	return webapi.NewClient(mockedServerUrl, testDeviceIdentifer, webapi.WithCustomClient(client), webapi.DisableRetries())
}
