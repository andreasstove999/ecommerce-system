import { createTheme } from '@mui/material/styles';
import { palette } from './palette';
import { typography } from './typography';
import { components } from './components';

export const appTheme = createTheme({
  palette,
  typography,
  components,
  shape: {
    borderRadius: 10,
  },
  spacing: 8,
});
