package ledger

import (
	"time"

	accounts "github.com/dani-dzotsi/double-entry-ledger/internal/accounts"
	constants "github.com/dani-dzotsi/double-entry-ledger/internal/platform"
	"github.com/google/uuid"
)

type Direction string

const (
	DEBIT Direction = "direct"
	CREDIT Direction ="credit"
)
type LedgerEntry struct{
	ID uuid.UUID `gorm:"type:uuid; primaryKey; gen_random_uuid()"`
	TransferID uuid.UUID `gorm:"type:uuid: not null" `
	Transfer Transfer `gorm:"foreignKey:TransaferID; references:ID; constraint:ondelete:RESTRICT,update:CASCADE"`
	AccountID uuid.UUID `gorm:"type:uuid; not null"`
	Account accounts.Account `gorm:"type:foreignKey;AccountID; references:ID; constraint:ondelete:RESTRICT,update:CASCADE"`
	Direction Direction `gorm:"type:text; check:chk_ledger_enties_direction, direction IN ('debit', 'credit'); not null"`
	Amount int64 `gorm:"type:bigint; not null; default=0"`
	Currency string `gorm:"type:char(3), not null"`
	balance_after int64 `gorm:"type:bigint; not null; check: chk_ledger_entries_balance_after, balance_after >= 0"`
	Created_at time.Time
}

type TransferType string

const (
	TRANSFER TransferType = "transfer"
	DEPOSIT TransferType = "deposit"
	REVERSAL TransferType ="reversal"
)

type TransferStatus string
const (
	POSTED TransferStatus = "posted"
	PENDING TransferStatus = "pending"
	REVERSED TransferStatus = "reversed"
	FAILED TransferStatus ="failed"

)


type Transfer struct {
	ID uuid.UUID `gorm:"type:uuid; primaryKey; default:gen_random_uuid()"`
	Type TransferType `gorm:"type:text; check:chk_transfers_type, type IN ('transfer', 'deposit', 'reversal')"`
	Status TransferStatus `gorm:"type:text; check:chk_transfers_status, status IN ('posted', 'failed', 'reversed', 'pending')"`
	IdempotencyKey string `gorm:"type:string; not null; unique"`
	RequestHash string `gorm:"type:text; not null; unique"` //hash of the request's normalized body, for conflict detection
	SourceAccountID uuid.UUID `gorm:"foreignKey:accounts, type:uuid; constraint:ondelete:RESTRICT, onupdate:CASCADE, not null"`
	DestAccountID uuid.UUID `gorm:"foreignKey:accounts, type:uuid; constraint:ondelete:RESTRICT, onupdate:CASCADE, not null"`
	SourceAmount int64 `gorm:"type:bigint; not null; check:chk_transfers_source_amount, source_amount > 0"`
	SourceCurrency constants.CurrencyTypes `gorm:"type:char(3); not null"`

}