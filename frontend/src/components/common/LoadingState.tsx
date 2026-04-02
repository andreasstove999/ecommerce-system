import { CircularProgress, Stack, Typography } from '@mui/material';

interface LoadingStateProps {
  label?: string;
}

export const LoadingState = ({ label = 'Loading...' }: LoadingStateProps) => (
  <Stack py={8} spacing={2} alignItems="center" justifyContent="center">
    <CircularProgress />
    <Typography color="text.secondary">{label}</Typography>
  </Stack>
);
