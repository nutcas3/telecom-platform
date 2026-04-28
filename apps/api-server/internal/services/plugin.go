package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// PluginService provides CRUD for installed plugins persisted in the database.
type PluginService struct {
	db *database.Database
}

func NewPluginService(db *database.Database) *PluginService {
	return &PluginService{db: db}
}

type PluginFilter struct {
	Enabled  *bool
	Type     string
	Category string
}

func (s *PluginService) List(ctx context.Context, f PluginFilter) ([]models.Plugin, error) {
	q := s.db.DB.WithContext(ctx).Model(&models.Plugin{})
	if f.Enabled != nil {
		q = q.Where("enabled = ?", *f.Enabled)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	var out []models.Plugin
	if err := q.Order("installed_at DESC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *PluginService) Get(ctx context.Context, id uint) (*models.Plugin, error) {
	var p models.Plugin
	if err := s.db.DB.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("plugin %d not found", id)
		}
		return nil, err
	}
	return &p, nil
}

type InstallPluginInput struct {
	Name        string `json:"name" binding:"required"`
	Version     string `json:"version" binding:"required"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	License     string `json:"license"`
	Homepage    string `json:"homepage"`
	Repository  string `json:"repository"`
	Config      string `json:"config,omitempty"` // JSON string
}

func (s *PluginService) Install(ctx context.Context, in InstallPluginInput) (*models.Plugin, error) {
	p := models.Plugin{
		Name:        in.Name,
		Version:     in.Version,
		Description: in.Description,
		Author:      in.Author,
		Type:        in.Type,
		Category:    in.Category,
		License:     in.License,
		Homepage:    in.Homepage,
		Repository:  in.Repository,
		Config:      in.Config,
		Enabled:     true,
		Status:      "installed",
		InstalledAt: time.Now(),
	}
	if err := s.db.DB.WithContext(ctx).Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PluginService) Uninstall(ctx context.Context, id uint) error {
	res := s.db.DB.WithContext(ctx).Delete(&models.Plugin{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("plugin %d not found", id)
	}
	return nil
}

func (s *PluginService) SetEnabled(ctx context.Context, id uint, enabled bool) (*models.Plugin, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Enabled = enabled
	if enabled {
		p.Status = "active"
	} else {
		p.Status = "inactive"
	}
	if err := s.db.DB.WithContext(ctx).Save(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}
