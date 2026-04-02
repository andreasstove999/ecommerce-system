import { Alert, Button, Stack } from '@mui/material';

interface ErrorStateProps {
  message: string;
  onRetry?: () => void;
}

export const ErrorState = ({ message, onRetry }: ErrorStateProps) => (
  <Stack py={2} spacing={2}>
    <Alert severity="error">{message}</Alert>
    {onRetry ? (
      <Stack alignItems="flex-start">
        <Button onClick={onRetry}>Try again</Button>
      </Stack>
    ) : null}
  </Stack>
);
