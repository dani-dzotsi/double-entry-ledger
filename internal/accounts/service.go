package accounts

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	GetAllAccounts(ctx context.Context, filter AccountListFilter) ([]AccountsResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func _serialize_account(account Account) AccountsResponse {
	return AccountsResponse{
		Id:               account.ID,
		Name:             account.Name,
		Currency:         account.CurrencyCode,
		AccountType:      account.Type,
		SystemRole:       account.SystemRole,
		CreatedAt:        account.CreatedAt,
		Balance:          account.Balance,
		HeldAmount:       account.HeldAmount,
		Status:           account.Status,
		AvailableBalance: int64(account.Balance - account.HeldAmount),
	}
}

func (s *service) GetAllAccounts(ctx context.Context, filter AccountListFilter) ([]AccountsResponse, error) {
	// get all the accounts
	results, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var accounts []AccountsResponse
	for _, value := range results {
		response := _serialize_account(value)
		accounts = append(accounts, response)
	}
	return accounts, nil
}

func (s *service) GetAccountByID(ctx context.Context, id uuid.UUID) (AccountsResponse, error) {
	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return AccountsResponse{}, err
	}

	return _serialize_account(result), nil
}

func (s *service) UpdateAccount(ctx context.Context, params)