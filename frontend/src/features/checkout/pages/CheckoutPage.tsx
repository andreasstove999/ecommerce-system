import { PageHeader } from '../../../components/common/PageHeader';
import { CheckoutForm } from '../components/CheckoutForm';

export const CheckoutPage = () => {
  return (
    <>
      <PageHeader title="Checkout" subtitle="Enter your shipping details and place your order." />
      <CheckoutForm />
    </>
  );
};
