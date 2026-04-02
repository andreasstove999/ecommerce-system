import { Stack, Typography } from '@mui/material';
import type { PropsWithChildren } from 'react';

interface AppSectionProps {
  title?: string;
  subtitle?: string;
}

export const AppSection = ({ title, subtitle, children }: PropsWithChildren<AppSectionProps>) => {
  return (
    <Stack spacing={2}>
      {title ? (
        <Stack spacing={0.5}>
          <Typography variant="h3">{title}</Typography>
          {subtitle ? <Typography color="text.secondary">{subtitle}</Typography> : null}
        </Stack>
      ) : null}
      {children}
    </Stack>
  );
};
