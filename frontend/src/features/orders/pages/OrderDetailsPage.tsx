import { useParams } from 'react-router-dom';
import { PageHeader } from '../../../components/common/PageHeader';
import { LoadingState } from '../../../components/common/LoadingState';
import { ErrorState } from '../../../components/common/ErrorState';
import { useOrder } from '../hooks/useOrder';
import { toErrorMessage } from '../../../utils/errors';
import { OrderSummary } from '../components/OrderSummary';

export const OrderDetailsPage = () => {
  const { id } = useParams();
  const orderQuery = useOrder(id);

  return (
    <>
      <PageHeader title="Order details" subtitle={`Order ID: ${id ?? 'Unknown'}`} />
      {orderQuery.isLoading ? <LoadingState label="Loading order..." /> : null}
      {orderQuery.isError ? <ErrorState message={toErrorMessage(orderQuery.error)} onRetry={() => void orderQuery.refetch()} /> : null}
      {orderQuery.data ? <OrderSummary order={orderQuery.data} /> : null}
    </>
  );
};
