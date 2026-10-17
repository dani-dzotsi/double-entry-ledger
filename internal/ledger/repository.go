package ledger

import (
	"context"
	"database/sql"
)


type LedgerEntryParams struct{
	DebitAccountId int64
	Amount int
	CreditAccountId int64
}

type DBTX interface{
	// ExecContext(context.Context, ...any) (sql.Result, error)
    // QueryContext(context.Context, string, ...any) (*sql.Rows, error)
    // QueryRowContext(context.Context, string, ...any) *sql.Row
	BeginTx(context.Context, *sql.TxOptions)(sql.Tx, error)
}
type Repository struct{}

type Repository interface{
		GetAccount(ctx context.Context, )
		LedgerEntry(ctx context.Context, params LedgerEntryParams)(error)
}

func NewRepository(db DBTX) Repository{
	return &repository{db: db}
}

func (r *repository) LedgerEntry(ctx context.Context, params LedgerEntryParams)(error){

	tx, err := r.db.BeginTx(ctx,nil)
	if err != nil{
		return err
	}
	defer tx.Rollback()

	const debitQuery = `
		UPDATE accounts SET amount = amount - @p1
		WHERE id = @p2
			AND amount >= @p1;
		`
	result, err := tx.ExecContext(ctx, debitQuery, params.Amount, params.DebitAccountId)
	if err != nil{
		return err
	}
	rows, err :=result.RowsAffected()

	if IsError(err){
		return err
	}

	if rows != 1{
		return sql.ErrNoRows
	}

	const creditQuery = `
		UPDATE accounts SET amount = amount + @p1
		WHERE ID = @p2;
	`
	result, err = tx.ExecContext(ctx, creditQuery, params.Amount, params.CreditAccountId)
	if err != nil{
		return err
	}

	if result!=1{
		return sql.ErrNoRows
	}
	return tx.Commit()
}

