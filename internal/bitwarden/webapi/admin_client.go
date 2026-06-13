package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
)

const adminCookieName = "VW_ADMIN"

// AdminClient talks to Vaultwarden's admin API (/admin/*), which is authenticated
// with an admin token exchanged for a session cookie. It is Vaultwarden-specific.
type AdminClient interface {
	CreateUser(ctx context.Context, email string) (*models.User, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	DeleteUser(ctx context.Context, userID string) error
}

type adminClient struct {
	serverURL  string
	adminToken string
	httpClient *http.Client
	loginMutex sync.Mutex
	loggedIn   bool
}

func NewAdminClient(serverURL, adminToken string) AdminClient {
	jar, _ := cookiejar.New(nil)
	return &adminClient{
		serverURL:  strings.TrimSuffix(serverURL, "/"),
		adminToken: adminToken,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}
}

func (c *adminClient) CreateUser(ctx context.Context, email string) (*models.User, error) {
	return adminRequest[models.User](ctx, c, func() (*http.Request, error) {
		return c.newRequest(ctx, "POST", c.serverURL+"/admin/invite", models.User{Email: email})
	})
}

func (c *adminClient) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return adminRequest[models.User](ctx, c, func() (*http.Request, error) {
		return c.newRequest(ctx, "GET", fmt.Sprintf("%s/admin/users/%s", c.serverURL, userID), nil)
	})
}

func (c *adminClient) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return adminRequest[models.User](ctx, c, func() (*http.Request, error) {
		return c.newRequest(ctx, "GET", fmt.Sprintf("%s/admin/users/by-mail/%s", c.serverURL, url.PathEscape(email)), nil)
	})
}

func (c *adminClient) DeleteUser(ctx context.Context, userID string) error {
	// The endpoint matches on a JSON content-type even though it ignores the body.
	_, err := adminRequest[[]byte](ctx, c, func() (*http.Request, error) {
		return c.newRequest(ctx, "POST", fmt.Sprintf("%s/admin/users/%s/delete", c.serverURL, userID), struct{}{})
	})
	return err
}

// adminRequest authenticates if necessary, runs the request, and retries once after
// re-authenticating if the admin session has expired (HTTP 401).
func adminRequest[T any](ctx context.Context, c *adminClient, build func() (*http.Request, error)) (*T, error) {
	if err := c.ensureLoggedIn(ctx); err != nil {
		return nil, err
	}

	req, err := build()
	if err != nil {
		return nil, err
	}

	res, err := doRequest[T](ctx, c.httpClient, req)
	if httpErr, ok := IsHTTPError(err); ok && httpErr.GetStatusCode() == http.StatusUnauthorized {
		c.invalidateSession()
		if err := c.ensureLoggedIn(ctx); err != nil {
			return nil, err
		}
		req, err = build()
		if err != nil {
			return nil, err
		}
		return doRequest[T](ctx, c.httpClient, req)
	}
	return res, err
}

func (c *adminClient) newRequest(ctx context.Context, method, reqURL string, body interface{}) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshall admin request body: %w", err)
		}
		reader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
	if err != nil {
		return nil, fmt.Errorf("error preparing admin request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *adminClient) ensureLoggedIn(ctx context.Context) error {
	c.loginMutex.Lock()
	defer c.loginMutex.Unlock()

	if c.loggedIn {
		return nil
	}

	form := url.Values{"token": {c.adminToken}}
	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL+"/admin", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("error preparing admin login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error logging in to the admin API: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("admin login failed: invalid admin token")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("admin login failed with status %d", resp.StatusCode)
	}

	adminURL, err := url.Parse(c.serverURL + "/admin")
	if err != nil {
		return fmt.Errorf("error parsing admin URL: %w", err)
	}
	for _, cookie := range c.httpClient.Jar.Cookies(adminURL) {
		if cookie.Name == adminCookieName {
			c.loggedIn = true
			return nil
		}
	}
	return fmt.Errorf("admin login did not return a session cookie")
}

func (c *adminClient) invalidateSession() {
	c.loginMutex.Lock()
	defer c.loginMutex.Unlock()
	c.loggedIn = false
}
