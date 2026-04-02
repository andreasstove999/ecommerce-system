export interface CartItem {
  id: string;
  productId: string;
  productName: string;
  unitPrice: number;
  quantity: number;
}

export interface Cart {
  id: string;
  items: CartItem[];
  totalItems: number;
  subtotal: number;
}

export interface AddToCartInput {
  productId: string;
  quantity: number;
  price: number;
}

export interface UpdateCartItemInput {
  quantity: number;
}
