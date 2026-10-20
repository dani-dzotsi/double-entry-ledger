package currencies

import "gorm.io/gorm"

import "strings"

type Currency struct{
	Code string `gorm:"type:char(3);primaryKey;check:chk_currencies_code_upper,code = upper(code)"`
	Name string `gorm:"type:text; not null"`
	Exponent int16 `gorm:"type:smallint; not null;check:chk_currencies_exponent_range, exponent BETWEEN 0 AND 6"`
	Active *bool `gorm:"default:true; not null"`
}


// Normalise before you save insert/update

func (c *Currency) BeforeSave(tx *gorm.DB){
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
}