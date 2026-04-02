import { useMutation, useQueryClient } from '@tanstack/react-query';
import { addToCart } from '../api/cartApi';
import { queryKeys } from '../../../lib/queryKeys';

export const useAddToCart = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: addToCart,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.cart });
    },
  });
};
