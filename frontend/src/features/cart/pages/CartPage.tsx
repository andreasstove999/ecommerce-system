import { Grid, Stack } from '@mui/material';
import { PageHeader } from '../../../components/common/PageHeader';
import { LoadingState } from '../../../components/common/LoadingState';
import { ErrorState } from '../../../components/common/ErrorState';
import { EmptyState } from '../../../components/common/EmptyState';
import { useCart } from '../hooks/useCart';
import { toErrorMessage } from '../../../utils/errors';
import { AppCard } from '../../../components/ui/AppCard';
import { CartItemRow } from '../components/CartItemRow';
import { CartSummary } from '../components/CartSummary';
import { useUpdateCartItem } from '../hooks/useUpdateCartItem';

export const CartPage = () => {
  const cartQuery = useCart();
  const updateItemMutation = useUpdateCartItem();

  const cart = cartQuery.data;

  return (
    <>
      <PageHeader title="Your cart" subtitle="Review and update items before checkout." />
      {cartQuery.isLoading ? <LoadingState label="Loading cart..." /> : null}
      {cartQuery.isError ? <ErrorState message={toErrorMessage(cartQuery.error)} onRetry={() => void cartQuery.refetch()} /> : null}
      {!cartQuery.isLoading && !cartQuery.isError && cart && cart.items.length === 0 ? (
        <EmptyState title="Your cart is empty" description="Add some products to begin checkout." />
      ) : null}
      {cart && cart.items.length > 0 ? (
        <Grid container spacing={3}>
          <Grid size={{ xs: 12, md: 8 }}>
            <AppCard>
              <Stack>
                {cart.items.map((item) => (
                  <CartItemRow
                    key={item.id}
                    item={item}
                    onQuantityChange={(itemId, quantity) =>
                      updateItemMutation.mutate({
                        itemId,
                        input: { quantity: Math.max(1, quantity) },
                      })
                    }
                  />
                ))}
              </Stack>
            </AppCard>
          </Grid>
          <Grid size={{ xs: 12, md: 4 }}>
            <CartSummary cart={cart} />
          </Grid>
        </Grid>
      ) : null}
    </>
  );
};
