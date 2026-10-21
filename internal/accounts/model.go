package accounts

import (
	"time"

	"github.com/dani-dzotsi/double-entry-ledger/internal/currencies"
	"github.com/google/uuid"
)

type AccountType string
type SystemRole string
type AccountStatus string

const (
	CUSTOMER AccountType = "CUSTOMER"
	SYSTEM   AccountType = "SYSTEM"
)

const (
	FUNDING     SystemRole = "FUNDING"
	FX_CLEARING SystemRole = "FX_CLEARING"
	ROUNDING    SystemRole = "ROUNDING"
)

const (
	ACTIVE AccountStatus = "ACTIVE"
	FROZEN AccountStatus = "FROZEN"
	CLOSED AccountStatus = "CLOSED"
)

type Account struct {
	ID           uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string              `gorm:"type:text;not null"`
	Type         AccountType         `gorm:"type:text;not null;check:chk_accounts_type,type IN ('CUSTOMER','SYSTEM')"`
	SystemRole   *SystemRole         `gorm:"type:text;uniqueIndex:ux_accounts_system_role_currency,priority:1,where:system_role IS NOT NULL;check:chk_accounts_system_role,(type = 'SYSTEM' AND system_role IN ('FUNDING','FX_CLEARING','ROUNDING')) OR (type = 'CUSTOMER' AND system_role IS NULL)"`
	CurrencyCode string              `gorm:"column:currency;type:char(3);not null;uniqueIndex:ux_accounts_system_role_currency,priority:2,where:system_role IS NOT NULL"`
	Currency     currencies.Currency `gorm:"foreignKey:CurrencyCode;references:Code;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE"`
	Balance      int64               `gorm:"type:bigint;not null;default:0;check:chk_accounts_customer_available_balance,type = 'SYSTEM' OR balance - held_amount >= 0"`
	HeldAmount   int64               `gorm:"type:bigint;not null;default:0;check:chk_accounts_held_amount,held_amount >= 0"`
	Status       AccountStatus       `gorm:"type:text;not null;default:'ACTIVE';check:chk_accounts_status,status IN ('ACTIVE','FROZEN','CLOSED')"`
	CreatedAt    time.Time           `gorm:"type:timestamptz;not null"`
	UpdatedAt    time.Time           `gorm:"type:timestamptz;not null"`
}
