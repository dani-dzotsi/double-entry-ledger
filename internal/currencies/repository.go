package currencies

import (
	"context"

	"gorm.io/gorm"
)

// Repository reads the currencies available to API clients. Currency data is
// provisioned by versioned migrations, so there is no runtime create method.
type Repository interface {
	ListActive(ctx context.Context) ([]Currency, error)
	GetActiveByCode(ctx context.Context, code string) (Currency, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListActive(ctx context.Context) ([]Currency, error) {
	var currencies []Currency
	err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Order("code ASC").
		Find(&currencies).Error
	if err != nil {
		return nil, err
	}
	return currencies, nil
}

func (r *repository) GetActiveByCode(ctx context.Context, code string) (Currency, error){
	// check if the currency with the pecified code is active and return theereesponse
	var currency Currency
	err := r.db.WithContext(ctx).Where("code = ? AND active = ?",code, true).First(&currency).Error
	return currency, err
}