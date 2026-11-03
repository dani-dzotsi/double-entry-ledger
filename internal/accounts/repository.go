package accounts

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountListFilter contains validated filters and the decoded keyset cursor.
// Cursor encoding/decoding and default/max limit validation belong in the API
// or service layer; the repository applies the resulting values to the query.
type AccountListFilter struct {
	Currency   string
	Status     AccountStatus
	Type       AccountType
	Limit      int
	BeforeTime *time.Time
	BeforeID   *uuid.UUID
}

type Repository interface {
	List(ctx context.Context, filter AccountListFilter) ([]Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (Account, error)
	Create(ctx context.Context, account *Account) (Account, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status AccountStatus) (Account, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context, filter AccountListFilter) ([]Account, error) {
	query := r.db.WithContext(ctx).Model(&Account{})
	if filter.Currency != "" {
		query = query.Where("currency = ?", filter.Currency)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.BeforeTime != nil && filter.BeforeID != nil {
		query = query.Where("(created_at, id) < (?, ?)", *filter.BeforeTime, *filter.BeforeID)
	}

	query = query.Order("created_at DESC").Order("id DESC")
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	var accounts []Account
	if err := query.Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (Account, error) {
	var account Account
	err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error
	return account, err
}

func (r *repository) Create(ctx context.Context, account *Account) (Account, error) {
	if account == nil {
		return Account{}, gorm.ErrInvalidData
	}
	now := time.Now().UTC()
	if account.CreatedAt.IsZero() {
		account.CreatedAt = now
	}
	if account.UpdatedAt.IsZero() {
		account.UpdatedAt = account.CreatedAt
	}
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return Account{}, err
	}
	return *account, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status AccountStatus) (Account, error) {
	result := r.db.WithContext(ctx).
		Model(&Account{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return Account{}, result.Error
	}
	if result.RowsAffected == 0 {
		return Account{}, gorm.ErrRecordNotFound
	}
	return r.GetByID(ctx, id)
}
