package clients

import (
	"context"
	"io"
	"net/http"
)

type InventoryClient struct{ c *Client }

func NewInventoryClient(c *Client) *InventoryClient { return &InventoryClient{c: c} }

func (ic *InventoryClient) GetAvailability(ctx context.Context, productId, rawQuery string, headers http.Header) (*http.Response, error) {
	return ic.c.Do(ctx, http.MethodGet, "/api/inventory/"+productId, rawQuery, nil, headers)
}

func (ic *InventoryClient) Adjust(ctx context.Context, rawQuery string, body io.Reader, headers http.Header) (*http.Response, error) {
	return ic.c.Do(ctx, http.MethodPost, "/api/inventory/adjust", rawQuery, body, headers)
}
