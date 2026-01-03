package clients

import (
	"context"
	"net/http"
)

type OrderClient struct{ c *Client }

func NewOrderClient(c *Client) *OrderClient { return &OrderClient{c: c} }

func (oc *OrderClient) GetOrder(ctx context.Context, orderId, rawQuery string, headers http.Header) (*http.Response, error) {
	return oc.c.Do(ctx, http.MethodGet, "/api/orders/"+orderId, rawQuery, nil, headers)
}

func (oc *OrderClient) ListOrdersByUser(ctx context.Context, userId, rawQuery string, headers http.Header) (*http.Response, error) {
	return oc.c.Do(ctx, http.MethodGet, "/api/users/"+userId+"/orders", rawQuery, nil, headers)
}
