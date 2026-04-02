export interface OrderItem {
  id: string;
  productName: string;
  unitPrice: number;
  quantity: number;
}

export interface Order {
  id: string;
  status: string;
  createdAt: string;
  total: number;
  items: OrderItem[];
}
