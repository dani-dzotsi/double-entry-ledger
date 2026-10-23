package accounts

import (
	"context"
	"errors"
	"strings"

	currency "github.com/dani-dzotsi/double-entry-ledger/internal/currencies"
	"github.com/google/uuid"
)

type Service interface {
	GetAllAccounts(ctx context.Context, filter AccountListFilter) ([]AccountsResponse, error)
	Create(ctx context.Context, params CreateAccountsRequest) (AccountsResponse, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (AccountsResponse, error)
	UpdateAccountStatus(ctx context.Context, id uuid.UUID, status UpdateAccountsStatusRequest) (AccountsResponse, error)
}

type service struct {
	repo         Repository
	currencyRepo currency.Repository
}

func NewService(repo Repository, currencyRepo currency.Repository) Service {
	return &service{repo: repo, currencyRepo: currencyRepo}
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

func (s *service) Create(ctx context.Context, params CreateAccountsRequest) (AccountsResponse, error) {
	//  create the new account
	// check if currency is active first
	curr := strings.TrimSpace(strings.ToUpper(params.Currency))
	_, err := s.currencyRepo.GetActiveByCode(ctx, curr)
	if err != nil {
		return AccountsResponse{}, err
	}
	account := &Account{
		Name:         params.Name,
		Type:         CUSTOMER,
		SystemRole:   nil,
		CurrencyCode: curr,
	}
	acc, err := s.repo.Create(ctx, account)
	if err != nil {
		return AccountsResponse{}, err
	}
	response := _serialize_account(acc)
	return response, nil
}

func _check_balanace_holding(account Account) error {
	// check that account.Balance is = 0
	if !(account.Balance == 0) {
		err := errors.New("account balance is not 0")
		return err
	}

	if !(account.HeldAmount == 0) {
		err := errors.New("account held balanace is not 0")
		return err
	}
	return nil

}
func (s *service) UpdateAccountStatus(ctx context.Context, id uuid.UUID, status UpdateAccountsStatusRequest) (AccountsResponse, error) {
	// check if account balance and and holdings is 0
	if status.Status == "CLOSED" {
		// THE ACCOUNT BY ID AND CHECK IF ITS BALANCE IS EQUAL TO 0
		check_acc_blc, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return AccountsResponse{}, err
		}
		err = _check_balanace_holding(check_acc_blc)
		if err != nil {
			return AccountsResponse{}, err
		}

	}
	sts := status.Status
	if sts != "ACTIVE" && sts != "FROZEN" && sts != "CLOSED"{
		return AccountsResponse{}, errors.New("Status must be of ['ACTIVE', 'FROZEN', 'CLOSED'")
	}
	// update the status
	acc, err := s.repo.UpdateStatus(ctx, id, status.Status)
	if err != nil {
		return AccountsResponse{}, err
	}

	// serialize account
	response := _serialize_account(acc)
	return response, nil
}
