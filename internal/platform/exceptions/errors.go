package exceptions

import "errors"



var (
	ErrNotFound             = errors.New("not found")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrDuplicateTransfer    = errors.New("duplicate idempotency key")
	ErrIdempotencyKeyReused = errors.New("idempotency key reused")
	ErrSameAccount          = errors.New("same account")
	ErrCurrencyMismatch     = errors.New("currency mismatch")
)

type ErrorFormat struct{
	Message string `json:"message"`
	Status_code int `json:"status_code"`
	Status string 	`json:"status"`
}

func EntityNotFound(msg string) ErrorFormat {
	return ErrorFormat{
		Message:     msg,
		Status_code: 404,
		Status:      "Not_Found",
	}
}



// error check
func IsError(err error) error{
	if err != nil{
		return err
	}
	return nil
}