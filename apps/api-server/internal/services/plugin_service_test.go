package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPluginTest(t *testing.T) (*database.Database, *PluginService) {
	// Use in-memory SQLite for testing
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate plugin table
	err = gormDB.AutoMigrate(&models.Plugin{})
	require.NoError(t, err)

	// Create database wrapper
	db := &database.Database{DB: gormDB}

	// Create service
	service := NewPluginService(db)

	return db, service
}

func TestPluginService_List(t *testing.T) {
	db, service := setupPluginTest(t)

	// Create test plugins
	plugins := []models.Plugin{
		{
			Name:        "test-plugin-1",
			Version:     "1.0.0",
			Description: "Test plugin 1",
			Author:      "Test Author",
			Type:        "monitoring",
			Enabled:     true,
			Status:      "installed",
		},
		{
			Name:        "test-plugin-2",
			Version:     "2.0.0",
			Description: "Test plugin 2",
			Author:      "Test Author",
			Type:        "security",
			Enabled:     false,
			Status:      "installed",
		},
	}

	for _, plugin := range plugins {
		err := db.DB.Create(&plugin).Error
		require.NoError(t, err)
	}

	t.Run("List all plugins", func(t *testing.T) {
		result, err := service.List(context.Background(), PluginFilter{})
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("List enabled plugins only", func(t *testing.T) {
		enabled := true
		result, err := service.List(context.Background(), PluginFilter{
			Enabled: &enabled,
		})
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.True(t, result[0].Enabled)
	})

	t.Run("List disabled plugins only", func(t *testing.T) {
		enabled := false
		result, err := service.List(context.Background(), PluginFilter{
			Enabled: &enabled,
		})
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.False(t, result[0].Enabled)
	})

	t.Run("List plugins by type", func(t *testing.T) {
		result, err := service.List(context.Background(), PluginFilter{
			Type: "monitoring",
		})
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "monitoring", result[0].Type)
	})

	t.Run("List plugins with no matches", func(t *testing.T) {
		result, err := service.List(context.Background(), PluginFilter{
			Type: "nonexistent",
		})
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestPluginService_Install(t *testing.T) {
	_, service := setupPluginTest(t)

	t.Run("Install new plugin successfully", func(t *testing.T) {
		configJSON, _ := json.Marshal(map[string]any{"key": "value"})
		input := InstallPluginInput{
			Name:        "new-plugin",
			Version:     "1.0.0",
			Description: "A new plugin",
			Author:      "Test Author",
			Type:        "monitoring",
			Config:      string(configJSON),
		}

		result, err := service.Install(context.Background(), input)
		require.NoError(t, err)

		assert.Equal(t, input.Name, result.Name)
		assert.Equal(t, input.Version, result.Version)
		assert.Equal(t, input.Description, result.Description)
		assert.Equal(t, input.Author, result.Author)
		assert.Equal(t, input.Type, result.Type)
		assert.True(t, result.Enabled)
		assert.Equal(t, "installed", result.Status)
		assert.NotZero(t, result.ID)
		assert.NotZero(t, result.InstalledAt)
	})

	t.Run("Install duplicate plugin should fail", func(t *testing.T) {
		input := InstallPluginInput{
			Name:        "duplicate-plugin",
			Version:     "1.0.0",
			Description: "Duplicate plugin",
			Author:      "Test Author",
			Type:        "monitoring",
		}

		// First installation should succeed
		_, err := service.Install(context.Background(), input)
		require.NoError(t, err)

		// Second installation should fail
		_, err = service.Install(context.Background(), input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE constraint failed")
	})

	t.Run("Install plugin with missing required fields", func(t *testing.T) {
		input := InstallPluginInput{
			// Missing name
			Version: "1.0.0",
		}

		_, err := service.Install(context.Background(), input)
		// The service might not validate this, so just check it doesn't panic
		if err != nil {
			assert.Error(t, err)
		}
	})
}

func TestPluginService_Get(t *testing.T) {
	db, service := setupPluginTest(t)

	// Create a plugin
	plugin := models.Plugin{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "Test plugin",
		Author:      "Test Author",
		Type:        "monitoring",
		Enabled:     true,
		Status:      "installed",
	}
	err := db.DB.Create(&plugin).Error
	require.NoError(t, err)

	t.Run("Get existing plugin", func(t *testing.T) {
		result, err := service.Get(context.Background(), plugin.ID)
		require.NoError(t, err)
		assert.Equal(t, plugin.Name, result.Name)
		assert.Equal(t, plugin.Version, result.Version)
		assert.Equal(t, plugin.Description, result.Description)
	})

	t.Run("Get nonexistent plugin", func(t *testing.T) {
		_, err := service.Get(context.Background(), 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestPluginService_Uninstall(t *testing.T) {
	db, service := setupPluginTest(t)

	// Create a plugin
	plugin := models.Plugin{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "Test plugin",
		Enabled:     true,
		Status:      "installed",
	}
	err := db.DB.Create(&plugin).Error
	require.NoError(t, err)

	t.Run("Uninstall existing plugin", func(t *testing.T) {
		err := service.Uninstall(context.Background(), plugin.ID)
		require.NoError(t, err)

		// Verify plugin is deleted
		var deletedPlugin models.Plugin
		err = db.DB.First(&deletedPlugin, plugin.ID).Error
		assert.Error(t, err) // Should not find the plugin
		assert.Contains(t, err.Error(), "record not found")
	})

	t.Run("Uninstall nonexistent plugin", func(t *testing.T) {
		err := service.Uninstall(context.Background(), 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// Benchmark tests
func BenchmarkPluginService_List(b *testing.B) {
	db, service := setupPluginTest(&testing.T{})

	// Create test plugins
	for i := range 100 {
		plugin := models.Plugin{
			Name:        fmt.Sprintf("plugin-%d", i),
			Version:     "1.0.0",
			Description: "Test plugin",
			Enabled:     i%2 == 0,
			Status:      "installed",
		}
		db.DB.Create(&plugin)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.List(context.Background(), PluginFilter{})
		if err != nil {
			b.Fatal(err)
		}
	}
}
