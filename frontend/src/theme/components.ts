import type { Components } from '@mui/material/styles';

export const components: Components = {
  MuiButton: {
    defaultProps: {
      variant: 'contained',
      disableElevation: true,
    },
    styleOverrides: {
      root: {
        borderRadius: 10,
        textTransform: 'none',
        fontWeight: 600,
      },
    },
  },
  MuiTextField: {
    defaultProps: {
      size: 'small',
      fullWidth: true,
    },
  },
  MuiCard: {
    styleOverrides: {
      root: {
        borderRadius: 14,
      },
    },
  },
  MuiAppBar: {
    styleOverrides: {
      root: {
        boxShadow: 'none',
        borderBottom: '1px solid rgba(15, 23, 42, 0.12)',
      },
    },
  },
};
