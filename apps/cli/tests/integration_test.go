package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/app"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/commands"
	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// IntegrationTestSuite runs comprehensive integration tests for the CLI
type IntegrationTestSuite struct {
	suite.Suite
	server    *httptest.Server
	configDir string
	config    *types.CLIConfig
	cli       *app.EnhancedCLI
}

// SetupSuite sets up the test environment
func (suite *IntegrationTestSuite) SetupSuite() {
	// Create temporary config directory
	tempDir, err := os.MkdirTemp("", "cli-test")
	require.NoError(suite.T(), err)
	suite.configDir = tempDir

	// Create mock API server
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.handleMockAPI(w, r)
	}))

	// Create test configuration
	suite.config = &types.CLIConfig{
		APIEndpoint: suite.server.URL,
		APIToken:    "test-token",
		Profile:     "test",
		NoColor:     true,
		Verbose:     false,
		Theme:       "default",
	}

	// Create CLI instance
	suite.cli = app.NewEnhancedCLI()
}

// TearDownSuite cleans up the test environment
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.configDir != "" {
		os.RemoveAll(suite.configDir)
	}
}

// handleMockAPI handles mock API responses
func (suite *IntegrationTestSuite) handleMockAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/api/v1/subscribers":
		switch r.Method {
		case "GET":
			subscribers := []map[string]any{
				{
					"id":      1,
					"name":    "John Doe",
					"email":   "john@example.com",
					"phone":   "+1234567890",
					"status":  "active",
					"balance": 100.50,
				},
				{
					"id":      2,
					"name":    "Jane Smith",
					"email":   "jane@example.com",
					"phone":   "+0987654321",
					"status":  "inactive",
					"balance": 0.0,
				},
			}
			json.NewEncoder(w).Encode(map[string]any{
				"subscribers": subscribers,
				"total":       len(subscribers),
			})
		case "POST":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id":     3,
				"status": "created",
			})
		}
	case "/api/v1/services":
		services := []map[string]any{
			{
				"name":    "api-server",
				"status":  "running",
				"version": "v1.0.0",
				"cpu":     45.2,
				"memory":  "256MB",
				"uptime":  "2h15m",
			},
			{
				"name":    "charging-engine",
				"status":  "running",
				"version": "v1.0.0",
				"cpu":     23.8,
				"memory":  "128MB",
				"uptime":  "2h15m",
			},
		}
		json.NewEncoder(w).Encode(map[string]any{
			"services": services,
			"total":    len(services),
		})
	case "/api/v1/billing/invoices":
		invoices := []map[string]any{
			{
				"id":         "INV-001",
				"customer":   "John Doe",
				"amount":     100.50,
				"status":     "paid",
				"due_date":   "2023-12-31",
				"created_at": "2023-12-01",
			},
		}
		json.NewEncoder(w).Encode(map[string]any{
			"invoices": invoices,
			"total":    len(invoices),
		})
	case "/api/v1/health":
		json.NewEncoder(w).Encode(map[string]any{
			"status": "healthy",
			"services": map[string]string{
				"api-server":      "healthy",
				"charging-engine": "healthy",
				"packet-gateway":  "healthy",
			},
		})
	default:
		http.NotFound(w, r)
	}
}

// TestSubscribersCommand tests the subscribers command
func (suite *IntegrationTestSuite) TestSubscribersCommand() {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "list subscribers",
			args:     []string{"subscribers", "list"},
			expected: "John Doe",
		},
		{
			name:     "list subscribers with format",
			args:     []string{"subscribers", "list", "--format", "json"},
			expected: "John Doe",
		},
		{
			name:     "list enhanced subscribers",
			args:     []string{"subscribers", "list", "--enhanced"},
			expected: "Telecom Platform",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Capture stdout
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run command
			err := commands.HandleSubscribersEnhanced(tt.args[1:], suite.config)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			io.Copy(&buf, r)

			// Check results
			assert.NoError(suite.T(), err)
			assert.Contains(suite.T(), buf.String(), tt.expected)
		})
	}
}

// TestServicesCommand tests the services command
func (suite *IntegrationTestSuite) TestServicesCommand() {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "list services",
			args:     []string{"services", "list"},
			expected: "api-server",
		},
		{
			name:     "list services with format",
			args:     []string{"services", "list", "--format", "json"},
			expected: "api-server",
		},
		{
			name:     "list enhanced services",
			args:     []string{"services", "list", "--enhanced"},
			expected: "Telecom Platform",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Capture stdout
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run command
			err := commands.HandleServicesEnhanced(tt.args[1:], suite.config)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			io.Copy(&buf, r)

			// Check results
			assert.NoError(suite.T(), err)
			assert.Contains(suite.T(), buf.String(), tt.expected)
		})
	}
}

// TestBillingCommand tests the billing command
func (suite *IntegrationTestSuite) TestBillingCommand() {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "list invoices",
			args:     []string{"billing", "invoices"},
			expected: "Billing command: invoices",
		},
		{
			name:     "list invoices with format",
			args:     []string{"billing", "invoices", "--format", "json"},
			expected: "Billing command: invoices",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Capture stdout
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run command
			err := commands.HandleBillingEnhanced(tt.args[1:], suite.config)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			io.Copy(&buf, r)

			// Check results
			assert.NoError(suite.T(), err)
			assert.Contains(suite.T(), buf.String(), tt.expected)
		})
	}
}

// TestConfigCommand tests the config command
func (suite *IntegrationTestSuite) TestConfigCommand() {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "show config",
			args:     []string{"config", "show"},
			expected: "test",
		},
		{
			name:     "set config value",
			args:     []string{"config", "set", "timeout", "60"},
			expected: "",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Capture stdout
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run command
			err := commands.HandleConfigEnhanced(tt.args[1:], suite.config)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			io.Copy(&buf, r)

			// Check results
			assert.NoError(suite.T(), err)
			if tt.expected != "" {
				assert.Contains(suite.T(), buf.String(), tt.expected)
			}
		})
	}
}

// TestErrorHandling tests error handling scenarios
func (suite *IntegrationTestSuite) TestErrorHandling() {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "missing required argument",
			args:        []string{"subscribers", "create"},
			expectError: true,
		},
		{
			name:        "invalid command",
			args:        []string{"subscribers", "invalid-command"},
			expectError: true, // This returns an error for unknown command
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := commands.HandleSubscribersEnhanced(tt.args[1:], suite.config)
			if tt.expectError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

// TestConcurrency tests concurrent command execution
func (suite *IntegrationTestSuite) TestConcurrency() {
	// Run multiple commands concurrently
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results := make(chan error, 10)
	commandList := [][]string{
		{"subscribers", "list"},
		{"services", "list"},
		{"billing", "invoices"},
		{"config", "show"},
	}

	for _, cmd := range commandList {
		go func(args []string) {
			var err error
			switch args[0] {
			case "subscribers":
				err = commands.HandleSubscribersEnhanced(args[1:], suite.config)
			case "services":
				err = commands.HandleServicesEnhanced(args[1:], suite.config)
			case "billing":
				err = commands.HandleBillingEnhanced(args[1:], suite.config)
			case "config":
				err = commands.HandleConfigEnhanced(args[1:], suite.config)
			}

			select {
			case results <- err:
			case <-ctx.Done():
				results <- ctx.Err()
			}
		}(cmd)
	}

	// Collect results
	for range commandList {
		select {
		case err := <-results:
			assert.NoError(suite.T(), err)
		case <-ctx.Done():
			suite.T().Fatal("Timeout waiting for concurrent commands")
		}
	}
}

// TestPerformance tests command performance
func (suite *IntegrationTestSuite) TestPerformance() {
	start := time.Now()
	err := commands.HandleSubscribersEnhanced([]string{"list"}, suite.config)
	duration := time.Since(start)

	assert.NoError(suite.T(), err)
	assert.Less(suite.T(), duration, 5*time.Second, "Command should complete within 5 seconds")
}

// TestCLIApp tests the CLI application
func (suite *IntegrationTestSuite) TestCLIApp() {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "run subscribers command",
			args:     []string{"cli", "subscribers", "list"},
			expected: "Telecom Platform",
		},
		{
			name:     "run services command",
			args:     []string{"cli", "services", "list"},
			expected: "Telecom Platform",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Capture stdout
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run CLI command
			err := suite.cli.Run(tt.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			io.Copy(&buf, r)

			// Check results
			assert.NoError(suite.T(), err)
			assert.Contains(suite.T(), buf.String(), tt.expected)
		})
	}
}

// Run the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// BenchmarkCommands benchmarks command performance
func BenchmarkCommands(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "cli-bench")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"subscribers": []map[string]any{},
			"total":       0,
		})
	}))
	defer server.Close()

	cfg := &types.CLIConfig{
		APIEndpoint: server.URL,
		APIToken:    "test-token",
		Profile:     "bench",
		NoColor:     true,
		Verbose:     false,
		Theme:       "default",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := commands.HandleSubscribersEnhanced([]string{"list"}, cfg)
		require.NoError(b, err)
	}
}
