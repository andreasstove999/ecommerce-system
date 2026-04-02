import { Button } from '@mui/material';
import { PageHeader } from '../../../components/common/PageHeader';
import { LoadingState } from '../../../components/common/LoadingState';
import { ErrorState } from '../../../components/common/ErrorState';
import { EmptyState } from '../../../components/common/EmptyState';
import { ProductGrid } from '../components/ProductGrid';
import { useProducts } from '../hooks/useProducts';
import { toErrorMessage } from '../../../utils/errors';

export const ProductListPage = () => {
  const { data, isLoading, isError, error, refetch } = useProducts();

  return (
    <>
      <PageHeader title="Products" subtitle="Browse all available catalog items." />
      {isLoading ? <LoadingState label="Loading products..." /> : null}
      {isError ? <ErrorState message={toErrorMessage(error)} onRetry={() => void refetch()} /> : null}
      {!isLoading && !isError && data?.length === 0 ? (
        <EmptyState
          title="No products available"
          description="Products will appear here once they are available in the catalog service."
          action={<Button onClick={() => void refetch()}>Refresh</Button>}
        />
      ) : null}
      {!isLoading && !isError && data && data.length > 0 ? <ProductGrid products={data} /> : null}
    </>
  );
};
