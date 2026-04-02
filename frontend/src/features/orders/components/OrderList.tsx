import { Paper, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';
import { Link } from 'react-router-dom';
import { formatCurrency, formatDate } from '../../../lib/formatters';
import type { Order } from '../types/order';

interface OrderListProps {
  orders: Order[];
}

export const OrderList = ({ orders }: OrderListProps) => {
  return (
    <Paper variant="outlined">
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Order ID</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Date</TableCell>
            <TableCell align="right">Total</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {orders.map((order) => (
            <TableRow key={order.id} hover>
              <TableCell>
                <Typography component={Link} to={`/orders/${order.id}`} sx={{ textDecoration: 'none' }}>
                  {order.id}
                </Typography>
              </TableCell>
              <TableCell>{order.status}</TableCell>
              <TableCell>{formatDate(order.createdAt)}</TableCell>
              <TableCell align="right">{formatCurrency(order.total)}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Paper>
  );
};
