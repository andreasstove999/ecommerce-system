import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '../../../lib/queryKeys';
import { getOrders } from '../api/ordersApi';

export const useOrders = () => {
  return useQuery({
    queryKey: queryKeys.orders,
    queryFn: getOrders,
  });
};
