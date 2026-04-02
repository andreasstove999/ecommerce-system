import type { TypographyOptions } from '@mui/material/styles/createTypography';

export const typography: TypographyOptions = {
  fontFamily: ['Inter', 'Roboto', 'Helvetica', 'Arial', 'sans-serif'].join(','),
  h1: { fontSize: '2rem', fontWeight: 700 },
  h2: { fontSize: '1.5rem', fontWeight: 700 },
  h3: { fontSize: '1.25rem', fontWeight: 600 },
  h4: { fontSize: '1.125rem', fontWeight: 600 },
  body1: { fontSize: '1rem' },
  body2: { fontSize: '0.9rem' },
};
