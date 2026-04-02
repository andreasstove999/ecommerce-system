import { Stack, Typography } from '@mui/material';
import { AppCard } from '../../../components/ui/AppCard';
import { formatCurrency, formatDate } from '../../../lib/formatters';
import type { Order } from '../types/order';

interface OrderSummaryProps {
  order: Order;
}

export const OrderSummary = ({ order }: OrderSummaryProps) => {
  return (
    <AppCard>
      <Stack spacing={1.5}>
        <Typography variant="h4">Order {order.id}</Typography>
        <Typography color="text.secondary">Placed on {formatDate(order.createdAt)}</Typography>
        <Typography>Status: {order.status}</Typography>
        <Typography fontWeight={700}>Total: {formatCurrency(order.total)}</Typography>
        {order.items.map((item) => (
          <Stack key={item.id} direction="row" justifyContent="space-between">
            <Typography>
              {item.productName} × {item.quantity}
            </Typography>
            <Typography>{formatCurrency(item.unitPrice * item.quantity)}</Typography>
          </Stack>
        ))}
      </Stack>
    </AppCard>
  );
};
