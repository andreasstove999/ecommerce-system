import { apiRequest } from '../../../lib/apiClient';
import type { AddToCartInput, Cart, UpdateCartItemInput } from '../types/cart';

export const getCart = async (): Promise<Cart> => {
  return apiRequest<Cart>('/cart');
};

export const addToCart = async (input: AddToCartInput): Promise<Cart> => {
  return apiRequest<Cart>('/cart/items', {
    method: 'POST',
    body: input,
  });
};

export const updateCartItem = async (itemId: string, input: UpdateCartItemInput): Promise<Cart> => {
  return apiRequest<Cart>(`/cart/items/${itemId}`, {
    method: 'PATCH',
    body: input,
  });
};
