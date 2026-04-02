export interface CheckoutInput {
  fullName: string;
  email: string;
  addressLine1: string;
  city: string;
  state: string;
  postalCode: string;
}

export interface CheckoutResult {
  orderId: string;
  status: string;
}
