package inventory

type StockItem struct {
	ProductID string `json:"productId"`
	Available int    `json:"available"`
}

type Line struct {
	ProductID string
	Quantity  int
}

type DepletedLine struct {
	ProductID  string
	Requested  int
	Available  int
}

type ReserveResult struct {
	Reserved []Line
	Depleted []DepletedLine
}
