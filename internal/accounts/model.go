package accounts

import "time"


type accounts struct{
	Id int64 `json:"id"`
	Name string `json:"name"`
	Currency string `json:"currency"`
	AccountType string `json:"account_type"`
	SystemRole string `json:"system_role"`
	CreatedAt time.Time `json:"created_at"`
}