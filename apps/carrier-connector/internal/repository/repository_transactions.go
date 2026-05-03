package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/carrier-connector/internal/currency"
)

// CreateTransaction creates a new transaction
func (r *GormRepository) CreateTransaction(ctx context.Context, transaction *currency.Transaction) error {
	model := r.transactionToModel(transaction)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.WithError(err).Error("Failed to create transaction")
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetTransaction retrieves a transaction by ID
func (r *GormRepository) GetTransaction(ctx context.Context, id string) (*currency.Transaction, error) {
	var model currency.TransactionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction not found: %s", id)
		}
		r.logger.WithError(err).Error("Failed to get transaction")
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return r.modelToTransaction(&model)
}

// UpdateTransaction updates an existing transaction
func (r *GormRepository) UpdateTransaction(ctx context.Context, transaction *currency.Transaction) error {
	model := r.transactionToModel(transaction)

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.WithError(err).Error("Failed to update transaction")
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// DeleteTransaction deletes a transaction
func (r *GormRepository) DeleteTransaction(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&currency.TransactionModel{}, "id = ?", id).Error; err != nil {
		r.logger.WithError(err).Error("Failed to delete transaction")
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}

// ListTransactions retrieves transactions based on filter
func (r *GormRepository) ListTransactions(ctx context.Context, filter *currency.TransactionFilter) ([]*currency.Transaction, error) {
	query := r.db.WithContext(ctx).Model(&currency.TransactionModel{})

	// Apply filters
	if filter.ProfileID != "" {
		query = query.Where("profile_id = ?", filter.ProfileID)
	}
	if filter.SubscriptionID != "" {
		query = query.Where("subscription_id = ?", filter.SubscriptionID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Currency != "" {
		query = query.Where("currency = ?", filter.Currency)
	}
	if filter.FromDate != nil {
		query = query.Where("created_at >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		query = query.Where("created_at <= ?", *filter.ToDate)
	}
	if filter.MinAmount > 0 {
		query = query.Where("amount >= ?", filter.MinAmount)
	}
	if filter.MaxAmount > 0 {
		query = query.Where("amount <= ?", filter.MaxAmount)
	}

	// Apply sorting
	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortOrder == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var models []currency.TransactionModel
	if err := query.Find(&models).Error; err != nil {
		r.logger.WithError(err).Error("Failed to list transactions")
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	transactions := make([]*currency.Transaction, 0, len(models))
	for _, model := range models {
		tx, err := r.modelToTransaction(&model)
		if err != nil {
			r.logger.WithError(err).Error("Failed to convert transaction model")
			continue
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// CountTransactions counts transactions based on filter
func (r *GormRepository) CountTransactions(ctx context.Context, filter *currency.TransactionFilter) (int, error) {
	query := r.db.WithContext(ctx).Model(&currency.TransactionModel{})

	// Apply filters
	if filter.ProfileID != "" {
		query = query.Where("profile_id = ?", filter.ProfileID)
	}
	if filter.SubscriptionID != "" {
		query = query.Where("subscription_id = ?", filter.SubscriptionID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Currency != "" {
		query = query.Where("currency = ?", filter.Currency)
	}
	if filter.FromDate != nil {
		query = query.Where("created_at >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		query = query.Where("created_at <= ?", *filter.ToDate)
	}
	if filter.MinAmount > 0 {
		query = query.Where("amount >= ?", filter.MinAmount)
	}
	if filter.MaxAmount > 0 {
		query = query.Where("amount <= ?", filter.MaxAmount)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		r.logger.WithError(err).Error("Failed to count transactions")
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return int(count), nil
}
