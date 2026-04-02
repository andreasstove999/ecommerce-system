import { useQuery } from '@tanstack/react-query';
import { queryKeys } from '../../../lib/queryKeys';
import { getCart } from '../api/cartApi';

export const useCart = () => {
  return useQuery({
    queryKey: queryKeys.cart,
    queryFn: getCart,
  });
};
