export const queryKeys = {
  products: ['products'] as const,
  product: (id: string) => ['products', id] as const,
  cart: ['cart'] as const,
  orders: ['orders'] as const,
  order: (id: string) => ['orders', id] as const,
};
