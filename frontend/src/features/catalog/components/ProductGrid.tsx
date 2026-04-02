import { Grid } from '@mui/material';
import type { Product } from '../types/product';
import { ProductCard } from './ProductCard';

interface ProductGridProps {
  products: Product[];
}

export const ProductGrid = ({ products }: ProductGridProps) => (
  <Grid container spacing={2.5}>
    {products.map((product) => (
      <Grid key={product.id} size={{ xs: 12, sm: 6, md: 4 }}>
        <ProductCard product={product} />
      </Grid>
    ))}
  </Grid>
);
