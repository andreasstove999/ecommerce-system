import { apiRequest } from '../../../lib/apiClient';
import type { Product } from '../types/product';

export const getProducts = async (): Promise<Product[]> => {
  return apiRequest<Product[]>('/products');
};

export const getProduct = async (id: string): Promise<Product> => {
  return apiRequest<Product>(`/products/${id}`);
};
