package accounts

import (
	"database/sql"
	"time"
)


type CreateAccountsRequest struct {
	Name string `json:"name" binding:"required"`
	Currency string `json:"currency" binding:"required"`
	AccountType string `json:"account_type" oneoff:"CUSTOMER SYSTEM"`
	SystemRole string `json:"system_role" oneoff:"EXTERNAL RX_CLEARING"`
}

type AccountsResponse struct {
	Id int64 `json:"id"`
	Name string `json:"name"`
	Currency string `json:"currency"`
	AccountType sql.NullString `json:"account_type"`
	SystemRole sql.NullString `json:"system_role"`
	CreatedAt time.Time `json:"created_at"`
}