import { Box, Container } from '@mui/material';
import { Outlet } from 'react-router-dom';
import { TopBar } from './TopBar';
import { Footer } from './Footer';

export const AppShell = () => {
  return (
    <Box minHeight="100vh" display="flex" flexDirection="column">
      <TopBar />
      <Container component="main" maxWidth="lg" sx={{ py: 4, flexGrow: 1 }}>
        <Outlet />
      </Container>
      <Footer />
    </Box>
  );
};
