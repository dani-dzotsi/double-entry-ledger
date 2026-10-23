package accounts

import (
	"time"

	"github.com/google/uuid"
)

type CreateAccountsRequest struct {
	Name     string `json:"name" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

type UpdateAccountsStatusRequest struct {
	Status AccountStatus `json:"status,omitempty"`
}

type AccountsResponse struct {
	Id               uuid.UUID     `json:"id"`
	Name             string        `json:"name"`
	Currency         string        `json:"currency"`
	AccountType      AccountType        `json:"account_type"`
	SystemRole       *SystemRole   `json:"system_role"`
	CreatedAt        time.Time     `json:"created_at"`
	HeldAmount       int64         `json:"held_amount"`
	Balance          int64         `json:"balance"`
	Status           AccountStatus `json:"status"`
	AvailableBalance int64         `json:"available_balance"`
}
