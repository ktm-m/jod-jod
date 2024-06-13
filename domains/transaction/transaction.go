package transaction

import (
	"gorm.io/gorm"
	"time"
)

type Transaction struct {
	gorm.Model
	Date            time.Time `gorm:"type:timestamp; default:CURRENT_TIMESTAMP; column:date" json:"date"`
	Amount          float64   `gorm:"type:decimal(10,2); default:0.00; column:amount" json:"amount"`
	Category        string    `gorm:"type:varchar(50); default:'other'; column:category" json:"category"`
	TransactionType string    `gorm:"type:varchar(20); not null; column:transaction_type" json:"transaction_type"`
	Note            string    `gorm:"type:varchar(255); column:note" json:"note"`
	ImageUrl        string    `gorm:"type:varchar(255); column:image_url" json:"image_url"`
	SpenderId       int       `gorm:"type:int; not null; column:spender_id" json:"spender_id"`
}

type GetAllTxnFilter struct {
	Date     *time.Time `gorm:"column:date" query:"date"`
	Category string     `gorm:"column:category" query:"category"`
	TxnType  string     `gorm:"column:transaction_type" query:"transaction-type"`
}

type Pagination struct {
	PageItem int `query:"page-item"`
	Page     int `query:"page"`
}

type PeriodFilter struct {
	StartDate *time.Time `query:"start-date"`
	EndDate   *time.Time `query:"end-date"`
}

type GetByTxnTypeRequest struct {
	SpenderId uint   `gorm:"column:spender_id" json:"spender_id"`
	TxnType   string `gorm:"column:transaction_type" json:"txn_type"`
}

type GetByCategoryRequest struct {
	SpenderId uint   `gorm:"column:spender_id" json:"spender_id"`
	Category  string `gorm:"column:category" json:"category"`
	TxnType   string `gorm:"column:transaction_type" json:"transaction_type"`
}

type GetAllResponse struct {
	ID              uint       `gorm:"column:id" json:"id"`
	Date            *time.Time `gorm:"column:date" json:"date"`
	Amount          float64    `gorm:"column:amount" json:"amount"`
	Category        string     `gorm:"column:category" json:"category"`
	ImageUrl        string     `gorm:"column:image_url" json:"image_url"`
	TransactionType string     `gorm:"column:transaction_type" json:"transaction_type"`
}

type GetAllByTxnTypeResponse struct {
	ID       uint       `gorm:"column:id" json:"id"`
	Date     *time.Time `gorm:"column:date" json:"date"`
	Amount   float64    `gorm:"column:amount" json:"amount"`
	Category string     `gorm:"column:category" json:"category"`
	ImageUrl string     `gorm:"column:image_url" json:"image_url"`
}

type GetSummaryResponse struct {
	TotalAmount   float64 `gorm:"column:total_amount" json:"total_amount"`
	AveragePerDay float64 `gorm:"column:average_amount_per_day" json:"average_amount_per_day"`
	TotalTxn      int     `gorm:"column:total_transaction" json:"total_transaction"`
}

type GetBalanceResponse struct {
	TotalAmountEarned float64 `json:"total_amount_earned"`
	TotalAmountSpent  float64 `json:"total_amount_spent"`
	TotalAmountSaved  float64 `json:"total_amount_saved"`
}

type GetByCategoryResponse struct {
	ID       uint      `gorm:"column: id" json:"id"`
	Date     time.Time `gorm:"column: date" json:"date"`
	Amount   float64   `gorm:"column: amount" json:"amount"`
	ImageUrl string    `gorm:"column: image_url" json:"image_url"`
}

type TextractResult struct {
	Category string
	Amount   float64
}
