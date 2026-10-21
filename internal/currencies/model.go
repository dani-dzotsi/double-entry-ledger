package currencies

import (
	"strings"

	"gorm.io/gorm"
)

type Currency struct {
	Code     string `gorm:"type:char(3);primaryKey;not null;check:chk_currencies_code_upper,code = upper(code)"`
	Name     string `gorm:"type:text;not null"`
	Exponent int16  `gorm:"type:smallint;not null;check:chk_currencies_exponent_range,exponent BETWEEN 0 AND 6"`
	Active   bool   `gorm:"not null;default:true"`
}

func (c *Currency) BeforeSave(tx *gorm.DB) error {
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
	return nil
}
