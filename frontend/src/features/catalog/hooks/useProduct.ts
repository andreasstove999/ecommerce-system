import { useQuery } from '@tanstack/react-query';
import { getProduct } from '../api/catalogApi';
import { queryKeys } from '../../../lib/queryKeys';

export const useProduct = (id: string | undefined) => {
  return useQuery({
    queryKey: queryKeys.product(id ?? ''),
    queryFn: () => getProduct(id as string),
    enabled: Boolean(id),
  });
};
