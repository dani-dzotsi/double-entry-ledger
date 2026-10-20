package accounts

import (
	"time"
	"github.com/google/uuid"

	currency "github.com/dani-dzotsi/double-entry-ledger/internal/currencies"
)

type AccountType string
type SystemRole string
type AccountStatus string

const (
	ACTIVE AccountStatus = "active"
	FROZEN AccountStatus = "frozen"
	CLOSED AccountStatus = "closed"
)

const (
	FUNDING SystemRole = "funding"
	FX_CLEARING SystemRole = "fx_clearing"
	ROUNDING SystemRole ="rounding"
)

const (
	CUSTOMER AccountType = "customer"
	SYSTEM AccountType = "system"
)

type Account struct{
	Id uuid.UUID `gorm:"type:uuid; primaryKey; default:gen_random_uuid()"`
	Name string `gorm:"type:text; not null"`
	CurrencyCode string `gorm:"type:char(3); not null; index; uniqueIndex:ux_accounts_system_role_currency,where:system_role IS NOT NULL"`
	Currency currency.Currency `gorm:"foreignKey:CurrencyCode;constraint:OnDelete:CASCADE"`
	Type AccountType `gorm:"type:text; not null; check:chk_account_type,type IN ('customer', 'system'); not null"`
	SystemRole *SystemRole `gorm:"type:text;uniqueIndex:ux_accounts_system_role_currency;check:chk_accounts_system_role,(type = 'system' AND system_role IN ('funding','fx_clearing','rounding')) OR (type = 'customer' AND system_role IS NULL)"`
	Balance uint64 `gorm:"type:bigint; not null; default:0; check:chk_accounts_amounts,held_amount >= 0 AND (type = 'system' OR (balance >= 0 AND held_amount <= balance))"`
	HeldAmount uint64 `gorm:"type:bigint; not null; default:0"`
	Status AccountStatus `gorm:"type:text; not null; check: chk_accounts_status,status IN ('active','frozen', 'closed')"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// DeletedAt gorm.DeletedAt `gorm:"index"`  //enablles soft delete
}
