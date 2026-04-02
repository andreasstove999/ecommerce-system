import { Stack, Typography } from '@mui/material';
import { AppCard } from '../../../components/ui/AppCard';
import { AppButton } from '../../../components/ui/AppButton';
import { formatCurrency } from '../../../lib/formatters';
import type { Cart } from '../types/cart';
import { Link } from 'react-router-dom';

interface CartSummaryProps {
  cart: Cart;
}

export const CartSummary = ({ cart }: CartSummaryProps) => {
  return (
    <AppCard>
      <Stack spacing={2}>
        <Typography variant="h4">Summary</Typography>
        <Stack direction="row" justifyContent="space-between">
          <Typography color="text.secondary">Items</Typography>
          <Typography>{cart.totalItems}</Typography>
        </Stack>
        <Stack direction="row" justifyContent="space-between">
          <Typography color="text.secondary">Subtotal</Typography>
          <Typography fontWeight={700}>{formatCurrency(cart.subtotal)}</Typography>
        </Stack>
        <AppButton component={Link} to="/checkout" fullWidth>
          Continue to checkout
        </AppButton>
      </Stack>
    </AppCard>
  );
};
