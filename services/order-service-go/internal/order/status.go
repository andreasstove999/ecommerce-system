package order

type Status string

const (
	StatusPending       Status = "pending"
	StatusPaymentFailed Status = "payment_failed"
	// To use this status, make sure the stock is reserved before the payment is processed, will be used in the future
	StatusStockReserved Status = "stock_reserved"
	StatusCompleted     Status = "completed"
	StatusCancelled     Status = "cancelled"
)
