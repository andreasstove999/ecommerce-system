import { Card, CardContent, type CardProps } from '@mui/material';
import type { PropsWithChildren } from 'react';

export const AppCard = ({ children, ...rest }: PropsWithChildren<CardProps>) => {
  return (
    <Card {...rest}>
      <CardContent>{children}</CardContent>
    </Card>
  );
};
