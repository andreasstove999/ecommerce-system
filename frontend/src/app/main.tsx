import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './App';
import { AppThemeProvider } from './providers/AppThemeProvider';
import { AppQueryProvider } from './providers/AppQueryProvider';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AppThemeProvider>
      <AppQueryProvider>
        <App />
      </AppQueryProvider>
    </AppThemeProvider>
  </StrictMode>,
);
