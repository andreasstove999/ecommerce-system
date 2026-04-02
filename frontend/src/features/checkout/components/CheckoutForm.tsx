import { Alert, Stack } from '@mui/material';
import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { AppButton } from '../../../components/ui/AppButton';
import { AppCard } from '../../../components/ui/AppCard';
import { AppTextField } from '../../../components/ui/AppTextField';
import { submitCheckout } from '../api/checkoutApi';
import { toErrorMessage } from '../../../utils/errors';
import type { CheckoutInput } from '../types/checkout';

const initialForm: CheckoutInput = {
  fullName: '',
  email: '',
  addressLine1: '',
  city: '',
  state: '',
  postalCode: '',
};

export const CheckoutForm = () => {
  const [formValues, setFormValues] = useState<CheckoutInput>(initialForm);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      const result = await submitCheckout(formValues);
      navigate(`/orders/${result.orderId}`);
    } catch (submitError) {
      setError(toErrorMessage(submitError));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AppCard>
      <Stack component="form" spacing={2} onSubmit={onSubmit}>
        {error ? <Alert severity="error">{error}</Alert> : null}
        <AppTextField label="Full name" required value={formValues.fullName} onChange={(e) => setFormValues({ ...formValues, fullName: e.target.value })} />
        <AppTextField label="Email" type="email" required value={formValues.email} onChange={(e) => setFormValues({ ...formValues, email: e.target.value })} />
        <AppTextField label="Address" required value={formValues.addressLine1} onChange={(e) => setFormValues({ ...formValues, addressLine1: e.target.value })} />
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
          <AppTextField label="City" required value={formValues.city} onChange={(e) => setFormValues({ ...formValues, city: e.target.value })} />
          <AppTextField label="State" required value={formValues.state} onChange={(e) => setFormValues({ ...formValues, state: e.target.value })} />
          <AppTextField label="Postal code" required value={formValues.postalCode} onChange={(e) => setFormValues({ ...formValues, postalCode: e.target.value })} />
        </Stack>
        <AppButton type="submit" disabled={isSubmitting}>
          {isSubmitting ? 'Placing order...' : 'Place order'}
        </AppButton>
      </Stack>
    </AppCard>
  );
};
