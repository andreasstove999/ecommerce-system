import { PageHeader } from '../../../components/common/PageHeader';
import { LoadingState } from '../../../components/common/LoadingState';
import { ErrorState } from '../../../components/common/ErrorState';
import { EmptyState } from '../../../components/common/EmptyState';
import { useOrders } from '../hooks/useOrders';
import { toErrorMessage } from '../../../utils/errors';
import { OrderList } from '../components/OrderList';

export const OrdersPage = () => {
  const ordersQuery = useOrders();

  return (
    <>
      <PageHeader title="Orders" subtitle="Track and review your previous orders." />
      {ordersQuery.isLoading ? <LoadingState label="Loading orders..." /> : null}
      {ordersQuery.isError ? (
        <ErrorState message={toErrorMessage(ordersQuery.error)} onRetry={() => void ordersQuery.refetch()} />
      ) : null}
      {ordersQuery.data && ordersQuery.data.length === 0 ? (
        <EmptyState title="No orders yet" description="Completed orders will show up here." />
      ) : null}
      {ordersQuery.data && ordersQuery.data.length > 0 ? <OrderList orders={ordersQuery.data} /> : null}
    </>
  );
};
