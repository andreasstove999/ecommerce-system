package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/andreasstove999/ecommerce-system/api-gateway-go/internal/middleware"
)

type Client struct {
	Name    string
	BaseURL *url.URL
	HTTP    *http.Client
}

func NewClient(name string, baseURL string, httpClient *http.Client) *Client {
	u, err := url.Parse(baseURL)
	if err != nil {
		// Fail fast: config error
		panic(fmt.Sprintf("invalid %s base url %q: %v", name, baseURL, err))
	}
	return &Client{Name: name, BaseURL: u, HTTP: httpClient}
}

func (c *Client) Do(ctx context.Context, method, path, rawQuery string, body io.Reader, inHeaders http.Header) (*http.Response, error) {
	rel := &url.URL{Path: path, RawQuery: rawQuery}
	u := c.BaseURL.ResolveReference(rel)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	copyHeaders(req.Header, inHeaders)

	// Ensure correlation id propagated downstream
	if cid := middleware.GetCorrelationID(ctx); cid != "" {
		req.Header.Set(middleware.HeaderCorrelationID, cid)
	}

	return c.HTTP.Do(req)
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		if isHopByHopHeader(k) {
			continue
		}
		// Host is not a header key here (it's req.Host), but keep this rule anyway
		if strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// Hop-by-hop headers (RFC 7230)
func isHopByHopHeader(k string) bool {
	switch http.CanonicalHeaderKey(k) {
	case "Connection", "Proxy-Connection", "Keep-Alive",
		"Proxy-Authenticate", "Proxy-Authorization",
		"Te", "Trailer", "Transfer-Encoding", "Upgrade":
		return true
	default:
		return false
	}
}
