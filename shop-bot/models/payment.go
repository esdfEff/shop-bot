package models

type Payment struct {
	ID        int
	UserID    int64
	InvoiceID int64
	Amount    float64
	Asset     string
	Status    string
	CreatedAt int64
	Purpose   string
}
