import { apiRequest } from '../../../lib/apiClient';
import type { CheckoutInput, CheckoutResult } from '../types/checkout';

export const submitCheckout = async (input: CheckoutInput): Promise<CheckoutResult> => {
  return apiRequest<CheckoutResult>('/me/cart/checkout', {
    method: 'POST',
    body: input,
  });
};
