import ShoppingCartIcon from '@mui/icons-material/ShoppingCart';
import { AppBar, Badge, Box, Button, Stack, Toolbar, Typography } from '@mui/material';
import { NavLink } from 'react-router-dom';
import { useCart } from '../../features/cart/hooks/useCart';

const links = [
  { to: '/products', label: 'Products' },
  { to: '/cart', label: 'Cart' },
  { to: '/orders', label: 'Orders' },
];

export const TopBar = () => {
  const { data: cart } = useCart();
  const count = cart?.items.reduce((sum, item) => sum + item.quantity, 0) ?? 0;

  return (
    <AppBar position="sticky" color="inherit">
      <Toolbar>
        <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ width: '100%' }}>
          <Stack direction="row" alignItems="center" spacing={3}>
            <Typography component={NavLink} to="/products" variant="h4" sx={{ textDecoration: 'none', color: 'inherit' }}>
              Commerce UI
            </Typography>
            <Stack direction="row" spacing={1}>
              {links.map((link) => (
                <Button
                  key={link.to}
                  component={NavLink}
                  to={link.to}
                  variant="text"
                  color="inherit"
                  sx={{
                    '&.active': {
                      color: 'primary.main',
                    },
                  }}
                >
                  {link.label}
                </Button>
              ))}
            </Stack>
          </Stack>

          <Box component={NavLink} to="/cart" sx={{ color: 'inherit', display: 'inline-flex' }}>
            <Badge badgeContent={count} color="primary">
              <ShoppingCartIcon />
            </Badge>
          </Box>
        </Stack>
      </Toolbar>
    </AppBar>
  );
};
