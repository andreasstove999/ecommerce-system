import { apiRequest } from '../../../lib/apiClient';
import type { Order } from '../types/order';

export const getOrders = async (): Promise<Order[]> => {
  return apiRequest<Order[]>('/me/orders');
};

export const getOrder = async (id: string): Promise<Order> => {
  return apiRequest<Order>(`/orders/${id}`);
};
