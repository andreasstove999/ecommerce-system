import { Navigate, createBrowserRouter } from 'react-router-dom';
import { AppShell } from '../components/layout/AppShell';
import { ProductListPage } from '../features/catalog/pages/ProductListPage';
import { ProductDetailsPage } from '../features/catalog/pages/ProductDetailsPage';
import { CartPage } from '../features/cart/pages/CartPage';
import { CheckoutPage } from '../features/checkout/pages/CheckoutPage';
import { OrdersPage } from '../features/orders/pages/OrdersPage';
import { OrderDetailsPage } from '../features/orders/pages/OrderDetailsPage';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppShell />,
    children: [
      { index: true, element: <Navigate to="/products" replace /> },
      { path: 'products', element: <ProductListPage /> },
      { path: 'products/:id', element: <ProductDetailsPage /> },
      { path: 'cart', element: <CartPage /> },
      { path: 'checkout', element: <CheckoutPage /> },
      { path: 'orders', element: <OrdersPage /> },
      { path: 'orders/:id', element: <OrderDetailsPage /> },
    ],
  },
]);
