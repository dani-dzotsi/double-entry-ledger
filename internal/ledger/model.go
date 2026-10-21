package ledger

import (
	"time"

	"github.com/dani-dzotsi/double-entry-ledger/internal/accounts"
	"github.com/dani-dzotsi/double-entry-ledger/internal/platform"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Direction string

const (
	DEBIT  Direction = "DEBIT"
	CREDIT Direction = "CREDIT"
)

type TransferType string

const (
	TRANSFER TransferType = "TRANSFER"
	DEPOSIT  TransferType = "DEPOSIT"
	REVERSAL TransferType = "REVERSAL"
)

type TransferStatus string

const (
	POSTED   TransferStatus = "POSTED"
	PENDING  TransferStatus = "PENDING"
	REVERSED TransferStatus = "REVERSED"
	FAILED   TransferStatus = "FAILED"
)

type FxRateSource string

const (
	MANUAL   FxRateSource = "MANUAL"
	PROVIDER FxRateSource = "PROVIDER"
)

// Transfer stores the lifecycle and amounts for a deposit, transfer, or reversal.
type Transfer struct {
	ID                uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type              TransferType           `gorm:"type:text;not null;check:chk_transfers_type,type IN ('TRANSFER','DEPOSIT','REVERSAL')"`
	Status            TransferStatus         `gorm:"type:text;not null;default:'PENDING';check:chk_transfers_status,status IN ('PENDING','POSTED','REVERSED','FAILED')"`
	IdempotencyKey    string                 `gorm:"type:text;not null;unique"`
	RequestHash       string                 `gorm:"type:text;not null"`
	SourceAccountID   uuid.UUID              `gorm:"type:uuid;not null;check:chk_transfers_distinct_accounts,source_account_id <> dest_account_id"`
	SourceAccount     accounts.Account       `gorm:"foreignKey:SourceAccountID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE"`
	DestAccountID     uuid.UUID              `gorm:"type:uuid;not null;check:chk_transfers_distinct_accounts,source_account_id <> dest_account_id"`
	DestAccount       accounts.Account       `gorm:"foreignKey:DestAccountID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE"`
	SourceAmount      int64                  `gorm:"type:bigint;not null;check:chk_transfers_source_amount,source_amount > 0"`
	SourceCurrency    platform.CurrencyTypes `gorm:"type:char(3);not null"`
	DestAmount        int64                  `gorm:"type:bigint;not null;check:chk_transfers_dest_amount,dest_amount > 0"`
	DestCurrency      platform.CurrencyTypes `gorm:"type:char(3);not null"`
	FxRateID          *uuid.UUID             `gorm:"type:uuid"`
	FxRate            *FxRate                `gorm:"foreignKey:FxRateID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE"`
	FxRoundingResidue *decimal.Decimal       `gorm:"type:numeric(18,8)"`
	ReversalOfID      *uuid.UUID             `gorm:"type:uuid;unique"`
	ReversalOf        *Transfer              `gorm:"foreignKey:ReversalOfID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE"`
	Memo              *string                `gorm:"type:text;check:chk_transfers_memo_length,memo IS NULL OR char_length(memo) <= 255"`
	FailureReason     *string                `gorm:"type:text"`
	CreatedAt         time.Time              `gorm:"type:timestamptz;not null"`
	UpdatedAt         time.Time              `gorm:"type:timestamptz;not null"`
	PostedAt          *time.Time             `gorm:"type:timestamptz"`
}

type LedgerEntry struct {
	ID           int64                  `gorm:"primaryKey;autoIncrement;type:bigint"`
	TransferID   uuid.UUID              `gorm:"type:uuid;not null;index"`
	Transfer     Transfer               `gorm:"foreignKey:TransferID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE"`
	AccountID    uuid.UUID              `gorm:"type:uuid;not null;index:idx_ledger_entries_account_id_id,priority:1"`
	Account      accounts.Account       `gorm:"foreignKey:AccountID;references:ID;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE"`
	Direction    Direction              `gorm:"type:text;not null;check:chk_ledger_entries_direction,direction IN ('DEBIT','CREDIT')"`
	Amount       int64                  `gorm:"type:bigint;not null;check:chk_ledger_entries_amount,amount > 0"`
	Currency     platform.CurrencyTypes `gorm:"type:char(3);not null"`
	BalanceAfter int64                  `gorm:"column:balance_after;type:bigint;not null"`
	CreatedAt    time.Time              `gorm:"type:timestamptz;not null;index:idx_ledger_entries_account_id_id,priority:2,sort:desc"`
}

type FxRate struct {
	ID            uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BaseCurrency  platform.CurrencyTypes `gorm:"type:char(3);not null;index:idx_fx_rates_pair_effective,priority:1"`
	QuoteCurrency platform.CurrencyTypes `gorm:"type:char(3);not null;index:idx_fx_rates_pair_effective,priority:2;check:chk_fx_rates_distinct_currencies,base_currency <> quote_currency"`
	Rate          decimal.Decimal        `gorm:"type:numeric(18,8);not null;check:chk_fx_rates_positive,rate > 0"`
	Source        FxRateSource           `gorm:"type:text;not null;check:chk_fx_rates_source,source IN ('MANUAL','PROVIDER')"`
	EffectiveAt   time.Time              `gorm:"type:timestamptz;not null;index:idx_fx_rates_pair_effective,priority:3,sort:desc"`
	CreatedAt     time.Time              `gorm:"type:timestamptz;not null"`
}
