package accounts

import (
	"context"
)


type AccountService interface{
	ListAccounts(ctx context.Context)([]AccountsResponse, error)
	CreateAccount(ctx context.Context, params CreateAccountParams)(AccountsResponse, error)
}
type service struct{
	repo AccountRepository
}

func NewService(repo AccountRepository) AccountService{
	return &service{repo:repo}
}


func (s *service) ListAccounts(ctx context.Context) ([]AccountsResponse, error){
	rows, err := s.repo.ListAccounts(ctx)
	if err != nil{
		return nil, err
	}
	response := make([]AccountsResponse, len(rows))
	for i,v := range rows{
		response[i] = AccountsResponse{
			Id: v.Id,
			Name : v.Name,
			Currency: v.Currency,		
			Account_type: v.Account_type,	
			System_role: v.System_role,
			CreatedAt: v.CreatedAt,			
		}
	}

	return response, nil
}

func (s *service) CreateAccount(ctx context.Context, params CreateAccountParams) (AccountsResponse, error){
	// create an account

	account, err := s.repo.CreateAccounts(ctx, params)
	if err != nil {
		return nil, err
	}
	
	response := AccountsResponse{
		Id: account.Id,
		Name: account.Name,
		Currency: account.Currency,
		AccountType :account.AccountType,
		SystemRole: account.SystemRole,
		CreatedAt: account.CreatedAt,
	}
	return response, nil
}
