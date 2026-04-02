import { Alert, Chip, Stack, Typography } from '@mui/material';
import { useParams } from 'react-router-dom';
import { PageHeader } from '../../../components/common/PageHeader';
import { LoadingState } from '../../../components/common/LoadingState';
import { ErrorState } from '../../../components/common/ErrorState';
import { AppButton } from '../../../components/ui/AppButton';
import { AppCard } from '../../../components/ui/AppCard';
import { formatCurrency } from '../../../lib/formatters';
import { useProduct } from '../hooks/useProduct';
import { useAddToCart } from '../../cart/hooks/useAddToCart';
import { toErrorMessage } from '../../../utils/errors';

export const ProductDetailsPage = () => {
  const { id } = useParams();
  const productQuery = useProduct(id);
  const addToCartMutation = useAddToCart();

  const product = productQuery.data;

  const handleAddToCart = () => {
    if (!product) {
      return;
    }

    addToCartMutation.mutate({ productId: product.id, quantity: 1 });
  };

  return (
    <>
      <PageHeader title="Product details" subtitle="View details and add item to cart." />
      {productQuery.isLoading ? <LoadingState /> : null}
      {productQuery.isError ? (
        <ErrorState message={toErrorMessage(productQuery.error)} onRetry={() => void productQuery.refetch()} />
      ) : null}
      {product ? (
        <AppCard>
          <Stack spacing={2}>
            <Stack direction="row" justifyContent="space-between" alignItems="center">
              <Typography variant="h3">{product.name}</Typography>
              {product.inStock ? <Chip color="success" label="In stock" /> : <Chip color="default" label="Out of stock" />}
            </Stack>
            <Typography color="text.secondary">{product.description}</Typography>
            <Typography fontWeight={700}>{formatCurrency(product.price)}</Typography>
            {addToCartMutation.isError ? <Alert severity="error">{toErrorMessage(addToCartMutation.error)}</Alert> : null}
            {addToCartMutation.isSuccess ? <Alert severity="success">Added to cart.</Alert> : null}
            <Stack direction="row" justifyContent="flex-start">
              <AppButton onClick={handleAddToCart} disabled={addToCartMutation.isPending || !product.inStock}>
                {addToCartMutation.isPending ? 'Adding...' : 'Add to cart'}
              </AppButton>
            </Stack>
          </Stack>
        </AppCard>
      ) : null}
    </>
  );
};
