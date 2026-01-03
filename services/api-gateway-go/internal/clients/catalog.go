package clients

import (
	"context"
	"io"
	"net/http"
)

type CatalogClient struct{ c *Client }

func NewCatalogClient(c *Client) *CatalogClient { return &CatalogClient{c: c} }

func (cc *CatalogClient) Health(ctx context.Context, rawQuery string, headers http.Header) (*http.Response, error) {
	return cc.c.Do(ctx, http.MethodGet, "/api/catalog/health", rawQuery, nil, headers)
}

func (cc *CatalogClient) ListProducts(ctx context.Context, rawQuery string, headers http.Header) (*http.Response, error) {
	return cc.c.Do(ctx, http.MethodGet, "/api/catalog/products", rawQuery, nil, headers)
}

func (cc *CatalogClient) GetProduct(ctx context.Context, id, rawQuery string, headers http.Header) (*http.Response, error) {
	return cc.c.Do(ctx, http.MethodGet, "/api/catalog/products/"+id, rawQuery, nil, headers)
}

func (cc *CatalogClient) CreateProduct(ctx context.Context, rawQuery string, body io.Reader, headers http.Header) (*http.Response, error) {
	return cc.c.Do(ctx, http.MethodPost, "/api/catalog/products", rawQuery, body, headers)
}
