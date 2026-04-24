package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// profileRow is the GORM-mapped row for the esim_profiles table.
type profileRow struct {
	ICCID          string    `gorm:"primaryKey;size:32"`
	EID            string    `gorm:"size:64;index"`
	IMSI           string    `gorm:"size:20;index"`
	MCC            string    `gorm:"size:4"`
	MNC            string    `gorm:"size:4"`
	ProfileType    string    `gorm:"size:32"`
	State          string    `gorm:"size:32;index"`
	TenantID       string    `gorm:"size:64;index"`
	ActivationCode string    `gorm:"size:255"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

// TableName sets the Postgres table name.
func (profileRow) TableName() string { return "esim_profiles" }

// PostgresProfileStore is a Postgres-backed ProfileRepository.
type PostgresProfileStore struct {
	db *gorm.DB
}

// NewPostgresProfileStore connects to Postgres using the given DSN, runs the
// profile schema migration, and returns a repository ready to use.
func NewPostgresProfileStore(dsn string) (*PostgresProfileStore, error) {
	if dsn == "" {
		return nil, errors.New("postgres DSN is empty")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.AutoMigrate(&profileRow{}); err != nil {
		return nil, fmt.Errorf("migrate esim_profiles: %w", err)
	}
	return &PostgresProfileStore{db: db}, nil
}

// Create inserts a profile row.
func (s *PostgresProfileStore) Create(ctx context.Context, p *Profile) error {
	row := toRow(p)
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	*p = *fromRow(row)
	return nil
}

// Get fetches a profile by ICCID.
func (s *PostgresProfileStore) Get(ctx context.Context, iccid string) (*Profile, error) {
	var row profileRow
	err := s.db.WithContext(ctx).First(&row, "iccid = ?", iccid).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromRow(&row), nil
}

// List returns a filtered, paginated page of profiles.
func (s *PostgresProfileStore) List(ctx context.Context, f ListFilter) ([]*Profile, int, error) {
	q := s.db.WithContext(ctx).Model(&profileRow{})
	if f.TenantID != "" {
		q = q.Where("tenant_id = ?", f.TenantID)
	}
	if f.State != "" {
		q = q.Where("state = ?", f.State)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}

	var rows []profileRow
	if err := q.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*Profile, len(rows))
	for i := range rows {
		result[i] = fromRow(&rows[i])
	}
	return result, int(total), nil
}

// UpdateState updates the state column and returns the refreshed profile.
func (s *PostgresProfileStore) UpdateState(ctx context.Context, iccid, state string) (*Profile, error) {
	res := s.db.WithContext(ctx).Model(&profileRow{}).
		Where("iccid = ?", iccid).
		Update("state", state)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return s.Get(ctx, iccid)
}

// Delete removes a profile row.
func (s *PostgresProfileStore) Delete(ctx context.Context, iccid string) error {
	res := s.db.WithContext(ctx).Delete(&profileRow{}, "iccid = ?", iccid)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Close releases the underlying database connection.
func (s *PostgresProfileStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// toRow maps domain -> row. Passthrough of CreatedAt/UpdatedAt lets GORM hooks
// populate them on first save.
func toRow(p *Profile) *profileRow {
	return &profileRow{
		ICCID:          p.ICCID,
		EID:            p.EID,
		IMSI:           p.IMSI,
		MCC:            p.MCC,
		MNC:            p.MNC,
		ProfileType:    p.ProfileType,
		State:          p.State,
		TenantID:       p.TenantID,
		ActivationCode: p.ActivationCode,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// fromRow maps row -> domain.
func fromRow(r *profileRow) *Profile {
	return &Profile{
		ICCID:          r.ICCID,
		EID:            r.EID,
		IMSI:           r.IMSI,
		MCC:            r.MCC,
		MNC:            r.MNC,
		ProfileType:    r.ProfileType,
		State:          r.State,
		TenantID:       r.TenantID,
		ActivationCode: r.ActivationCode,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}
