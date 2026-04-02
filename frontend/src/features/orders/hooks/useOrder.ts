import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '../../../lib/queryKeys';
import { getOrder } from '../api/ordersApi';

export const useOrder = (id: string | undefined) => {
  return useQuery({
    queryKey: queryKeys.order(id ?? ''),
    queryFn: () => getOrder(id as string),
    enabled: Boolean(id),
  });
};
