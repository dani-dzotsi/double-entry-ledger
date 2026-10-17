package accounts

import (
	"context"
	"database/sql"
	"errors"
	"time"

)


type Account struct{
	Id int64
	Name string
	Currency string
	AccountType sql.NullString
	SystemRole sql.NullString
	CreatedAt time.Time
}
type CreateAccountParams struct{
	Name string
	Currency string
	AccountType sql.NullString
	SystemRole sql.NullString
	CreatedAt time.Time
}

type UpdateAccountParams struct{
	Name sql.NullString
	Currency sql.NullString
	Accout_type sql.NullString
	SystemRole sql.NullString
}

type DepositAccountParams struct{
	Amount int
	Idempotency_key string
	Currency string
}

type WIthdrawalAccountParams struct{
	Amount int
	Idempotency_key string
	Currency string
}

type AccountRepository interface{
	ListAccounts(ctx context.Context) ([]Account, error)
	CreateAccounts(ctx context.Context, params CreateAccountParams) (Account, error)
	UpdateAccount(ctx context.Context,params UpdateAccountParams, id int64) (Account, error)
	GetAccountsDetail(ctx context.Context, id int64)(Account, error)
	AccountStatement(ctx context.Context, id int64) (Account, error)
	DepositExternal(ctx context.Context, id int64, params DepositAccountParams) (error)
	Withdrawal(ctx context.Context, id int64, params WIthdrawalAccountParams) (error)
}
type DBTX interface {
    ExecContext(context.Context, string, ...any) (sql.Result, error)
    QueryContext(context.Context, string, ...any) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...any) *sql.Row
}

type repository struct {
	db DBTX
}

func NewRepository(db DBTX) AccountRepository {
	return &repository{db: db}
}



// The scan helper
type scanner interface{
	Scan(dest ...any) error
}

func scanAccount(s scanner) (Account, error){
	var a Account
	err := s.Scan(&a.Id, &a.Name, &a.Currency, &a.AccountType, &a.SystemRole)
	return a, err
}

func (r *repository) ListAccounts(ctx context.Context)([]Account, error){
	const q = `SELECT * FROM accounts ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil{
		return nil, err
	}
	defer rows.Close()

	var out []Account
	for rows.Next(){
		u, err := scanAccount(rows)
		if err != nil{
			return nil, err
		}

		out = append(out, u)
	}

	return out, rows.Err()

}




