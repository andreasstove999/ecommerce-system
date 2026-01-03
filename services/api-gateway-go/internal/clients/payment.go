package clients

import (
	"context"
	"net/http"
)

type PaymentClient struct{ c *Client }

func NewPaymentClient(c *Client) *PaymentClient { return &PaymentClient{c: c} }

func (pc *PaymentClient) Health(ctx context.Context, rawQuery string, headers http.Header) (*http.Response, error) {
	return pc.c.Do(ctx, http.MethodGet, "/health", rawQuery, nil, headers)
}

func (pc *PaymentClient) ByOrder(ctx context.Context, orderId, rawQuery string, headers http.Header) (*http.Response, error) {
	return pc.c.Do(ctx, http.MethodGet, "/api/payments/by-order/"+orderId, rawQuery, nil, headers)
}
