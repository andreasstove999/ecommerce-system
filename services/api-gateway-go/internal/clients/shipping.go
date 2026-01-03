package clients

import (
	"context"
	"net/http"
)

type ShippingClient struct{ c *Client }

func NewShippingClient(c *Client) *ShippingClient { return &ShippingClient{c: c} }

func (sc *ShippingClient) GetByID(ctx context.Context, shippingId, rawQuery string, headers http.Header) (*http.Response, error) {
	return sc.c.Do(ctx, http.MethodGet, "/api/shipping/"+shippingId, rawQuery, nil, headers)
}

func (sc *ShippingClient) ByOrder(ctx context.Context, orderId, rawQuery string, headers http.Header) (*http.Response, error) {
	return sc.c.Do(ctx, http.MethodGet, "/api/shipping/by-order/"+orderId, rawQuery, nil, headers)
}
