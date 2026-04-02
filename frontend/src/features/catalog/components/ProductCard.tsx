import { CardActions, CardMedia, Stack, Typography } from '@mui/material';
import { Link } from 'react-router-dom';
import { formatCurrency } from '../../../lib/formatters';
import type { Product } from '../types/product';
import { AppButton } from '../../../components/ui/AppButton';
import { AppCard } from '../../../components/ui/AppCard';

interface ProductCardProps {
  product: Product;
}

export const ProductCard = ({ product }: ProductCardProps) => {
  return (
    <AppCard>
      {product.imageUrl ? (
        <CardMedia component="img" height="180" image={product.imageUrl} alt={product.name} sx={{ borderRadius: 1.5, mb: 2 }} />
      ) : null}
      <Stack spacing={1}>
        <Typography variant="h4">{product.name}</Typography>
        <Typography color="text.secondary" noWrap>
          {product.description}
        </Typography>
        <Typography fontWeight={700}>{formatCurrency(product.price)}</Typography>
      </Stack>
      <CardActions sx={{ px: 0, pt: 2 }}>
        <AppButton component={Link} to={`/products/${product.id}`}>
          View details
        </AppButton>
      </CardActions>
    </AppCard>
  );
};
