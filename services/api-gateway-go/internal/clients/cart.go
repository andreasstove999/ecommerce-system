package clients

import (
	"context"
	"io"
	"net/http"
)

type CartClient struct{ c *Client }

func NewCartClient(c *Client) *CartClient { return &CartClient{c: c} }

func (cc *CartClient) AddItem(ctx context.Context, userId, rawQuery string, body io.Reader, headers http.Header) (*http.Response, error) {
	return cc.c.Do(ctx, http.MethodPost, "/api/cart/"+userId+"/items", rawQuery, body, headers)
}

func (cc *CartClient) GetCart(ctx context.Context, userId, rawQuery string, headers http.Header) (*http.Response, error) {
	return cc.c.Do(ctx, http.MethodGet, "/api/cart/"+userId, rawQuery, nil, headers)
}

func (cc *CartClient) Checkout(ctx context.Context, userId, rawQuery string, body io.Reader, headers http.Header) (*http.Response, error) {
	return cc.c.Do(ctx, http.MethodPost, "/api/cart/"+userId+"/checkout", rawQuery, body, headers)
}
