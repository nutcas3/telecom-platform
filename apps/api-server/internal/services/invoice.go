package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nutcas3/telecom-platform/apps/api-server/internal/database"
	"github.com/nutcas3/telecom-platform/apps/api-server/internal/models"
)

// InvoiceService handles invoice generation and management
type InvoiceService struct {
	db *database.Database
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(db *database.Database) *InvoiceService {
	return &InvoiceService{db: db}
}

// GenerateMonthlyInvoice generates a monthly invoice for a subscriber
func (s *InvoiceService) GenerateMonthlyInvoice(ctx context.Context, subscriberID uint, billingPeriod string) (*models.Invoice, error) {
	// Get subscriber
	var subscriber models.Subscriber
	if err := s.db.DB.First(&subscriber, subscriberID).Error; err != nil {
		return nil, fmt.Errorf("subscriber not found: %w", err)
	}

	// Get usage events for the billing period
	startDate, err := parseBillingPeriod(billingPeriod)
	if err != nil {
		return nil, fmt.Errorf("invalid billing period: %w", err)
	}
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	var usageEvents []models.UsageEvent
	if err := s.db.DB.Where("subscriber_id = ? AND timestamp BETWEEN ? AND ?",
		subscriberID, startDate, endDate).Find(&usageEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch usage events: %w", err)
	}

	// Create invoice with existing model structure
	invoice := &models.Invoice{
		SubscriberID: subscriberID,
		DueDate:      time.Now().AddDate(0, 0, 30), // Due in 30 days
		Status:       models.InvoiceStatusDraft,
		Currency:     "USD",
		CreatedAt:    time.Now(),
	}

	// Add plan fee
	planFee := 25.00 // Default monthly fee
	invoice.LineItems = append(invoice.LineItems, models.InvoiceLineItem{
		Description: "Monthly Plan Fee",
		Quantity:    1,
		UnitPrice:   planFee,
		Amount:      planFee,
	})

	// Add usage charges
	totalUsageCost := 0.0
	for _, event := range usageEvents {
		totalUsageCost += event.Cost
	}

	if totalUsageCost > 0 {
		invoice.LineItems = append(invoice.LineItems, models.InvoiceLineItem{
			Description: "Usage Charges",
			Quantity:    1,
			UnitPrice:   totalUsageCost,
			Amount:      totalUsageCost,
		})
	}

	// Calculate total amount
	invoice.Amount = planFee + totalUsageCost

	// Save invoice
	if err := s.db.DB.Create(invoice).Error; err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Load subscriber relationship
	s.db.DB.Preload("Subscriber").First(invoice, invoice.ID)

	return invoice, nil
}

// SendInvoice sends an invoice to the subscriber
func (s *InvoiceService) SendInvoice(ctx context.Context, invoiceID uint) error {
	var invoice models.Invoice
	if err := s.db.DB.First(&invoice, invoiceID).Error; err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Update status to pending
	invoice.Status = models.InvoiceStatusPending
	if err := s.db.DB.Save(&invoice).Error; err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	// TODO: Send email notification
	// This would integrate with an email service

	return nil
}

// MarkInvoiceAsPaid marks an invoice as paid
func (s *InvoiceService) MarkInvoiceAsPaid(ctx context.Context, invoiceID uint, paymentMethod string) error {
	var invoice models.Invoice
	if err := s.db.DB.First(&invoice, invoiceID).Error; err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	invoice.Status = models.InvoiceStatusPaid
	// Update payment date if needed

	return s.db.DB.Save(&invoice).Error
}

// GetOverdueInvoices returns all overdue invoices
func (s *InvoiceService) GetOverdueInvoices(ctx context.Context) ([]models.Invoice, error) {
	var invoices []models.Invoice
	err := s.db.DB.Where("status = ? AND due_date < ?",
		models.InvoiceStatusPending, time.Now()).Find(&invoices).Error
	return invoices, err
}

// GetSubscriberInvoices returns all invoices for a subscriber
func (s *InvoiceService) GetSubscriberInvoices(ctx context.Context, subscriberID uint) ([]models.Invoice, error) {
	var invoices []models.Invoice
	err := s.db.DB.Where("subscriber_id = ?", subscriberID).
		Preload("LineItems").
		Order("created_at DESC").
		Find(&invoices).Error
	return invoices, err
}

// parseBillingPeriod parses a billing period string like "2024-01" into a time.Time
func parseBillingPeriod(period string) (time.Time, error) {
	if len(period) != 7 || period[4] != '-' {
		return time.Time{}, fmt.Errorf("invalid billing period format, expected YYYY-MM")
	}

	year := period[:4]
	month := period[5:]

	layout := "2006-01"
	return time.Parse(layout, fmt.Sprintf("%s-%s", year, month))
}

// InvoiceGenerator handles automated invoice generation
type InvoiceGenerator struct {
	invoiceService *InvoiceService
}

// NewInvoiceGenerator creates a new invoice generator
func NewInvoiceGenerator(invoiceService *InvoiceService) *InvoiceGenerator {
	return &InvoiceGenerator{invoiceService: invoiceService}
}

// GenerateAllMonthlyInvoices generates monthly invoices for all active subscribers
func (g *InvoiceGenerator) GenerateAllMonthlyInvoices(ctx context.Context, billingPeriod string) error {
	// Get all active subscribers
	var subscribers []models.Subscriber
	if err := g.invoiceService.db.DB.Where("status = ?", "active").Find(&subscribers).Error; err != nil {
		return fmt.Errorf("failed to fetch subscribers: %w", err)
	}

	var errors []error
	for _, subscriber := range subscribers {
		// Check if invoice already exists for this period by checking creation date
		startDate, _ := parseBillingPeriod(billingPeriod)
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

		var existingInvoice models.Invoice
		err := g.invoiceService.db.DB.Where("subscriber_id = ? AND created_at BETWEEN ? AND ?",
			subscriber.ID, startDate, endDate).First(&existingInvoice).Error

		if err == nil {
			// Invoice already exists, skip
			continue
		}

		// Generate new invoice
		_, err = g.invoiceService.GenerateMonthlyInvoice(ctx, subscriber.ID, billingPeriod)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to generate invoice for subscriber %d: %w", subscriber.ID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors while generating invoices", len(errors))
	}

	return nil
}
