import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateCartItem } from '../api/cartApi';
import { queryKeys } from '../../../lib/queryKeys';
import type { UpdateCartItemInput } from '../types/cart';

interface UpdateCartItemVars {
  itemId: string;
  input: UpdateCartItemInput;
}

export const useUpdateCartItem = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ itemId, input }: UpdateCartItemVars) => updateCartItem(itemId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.cart });
    },
  });
};
