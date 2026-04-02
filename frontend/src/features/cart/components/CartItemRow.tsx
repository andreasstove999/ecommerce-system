import { Stack, Typography } from '@mui/material';
import { formatCurrency } from '../../../lib/formatters';
import type { CartItem } from '../types/cart';
import { AppTextField } from '../../../components/ui/AppTextField';

interface CartItemRowProps {
  item: CartItem;
  onQuantityChange: (itemId: string, quantity: number) => void;
}

export const CartItemRow = ({ item, onQuantityChange }: CartItemRowProps) => {
  return (
    <Stack
      direction={{ xs: 'column', sm: 'row' }}
      spacing={2}
      alignItems={{ xs: 'flex-start', sm: 'center' }}
      justifyContent="space-between"
      py={1.5}
      borderBottom="1px solid"
      borderColor="divider"
    >
      <Stack>
        <Typography fontWeight={600}>{item.productName}</Typography>
        <Typography variant="body2" color="text.secondary">
          Unit price: {formatCurrency(item.unitPrice)}
        </Typography>
      </Stack>

      <Stack direction="row" alignItems="center" spacing={2}>
        <AppTextField
          type="number"
          label="Qty"
          value={item.quantity}
          onChange={(event) => onQuantityChange(item.id, Number(event.target.value || 1))}
          sx={{ width: 90 }}
          slotProps={{ htmlInput: { min: 1 } }}
        />
        <Typography minWidth={90} textAlign="right" fontWeight={600}>
          {formatCurrency(item.unitPrice * item.quantity)}
        </Typography>
      </Stack>
    </Stack>
  );
};
