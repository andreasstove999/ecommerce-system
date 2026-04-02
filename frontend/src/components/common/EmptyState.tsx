import { Paper, Stack, Typography } from '@mui/material';
import type { ReactNode } from 'react';

interface EmptyStateProps {
  title: string;
  description?: string;
  action?: ReactNode;
}

export const EmptyState = ({ title, description, action }: EmptyStateProps) => (
  <Paper variant="outlined" sx={{ p: 4 }}>
    <Stack spacing={1.5} alignItems="flex-start">
      <Typography variant="h4">{title}</Typography>
      {description ? <Typography color="text.secondary">{description}</Typography> : null}
      {action}
    </Stack>
  </Paper>
);
