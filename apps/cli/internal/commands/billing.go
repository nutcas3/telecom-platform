package commands

import (
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// HandleBilling is the entry point for billing commands.
func HandleBilling(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) == 0 {
		showBillingHelp(u)
		return nil
	}

	u.connectivityBanner()

	command := args[0]
	switch command {
	case "invoices":
		return handleInvoices(u)
	case "payments":
		return handlePayments(u)
	case "generate":
		return generateInvoice(u, args[1:])
	default:
		u.errorln("Unknown billing command: " + command)
		showBillingHelp(u)
		return fmt.Errorf("unknown command: %s", command)
	}
}

func showBillingHelp(u *uiContext) {
	u.header("Billing Management")
	u.muted("Usage: telecom-cli billing <command> [options]")
	fmt.Println()

	t := u.newTable()
	t.AddColumn("Command", 26, "left")
	t.AddColumn("Description", 40, "left")
	t.AddRow("invoices", "List all invoices")
	t.AddRow("payments", "List all payments")
	t.AddRow("generate <subscriber>", "Generate invoice for subscriber")
	fmt.Println(t.Render())
}

func handleInvoices(u *uiContext) error {
	u.header("Recent Invoices")

	invoices, err := u.client.GetInvoices()
	t := u.newTable()
	t.AddColumn("Invoice #", 14, "left")
	t.AddColumn("Date", 12, "left")
	t.AddColumn("Status", 10, "left")
	t.AddColumn("Amount", 10, "right")
	t.AddColumn("Subscriber", 22, "left")

	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		t.AddStyledRow(statusStyle("PAID").Style, "INV-000001", "2026-05-30", "Paid", "$45.67", "John Doe")
		t.AddStyledRow(statusStyle("PENDING").Style, "INV-000002", "2026-05-30", "Pending", "$123.45", "Jane Smith")
		t.AddStyledRow(statusStyle("OVERDUE").Style, "INV-000003", "2024-01-14", "Overdue", "$67.89", "Bob Johnson")
		fmt.Println(t.Render())
		return nil
	}

	for _, inv := range invoices {
		t.AddStyledRow(statusStyle(inv.Status).Style,
			inv.ID,
			inv.CreatedAt.Format("2006-01-02"),
			inv.Status,
			fmt.Sprintf("$%.2f", inv.Amount),
			fmt.Sprintf("%s %s", inv.Subscriber.FirstName, inv.Subscriber.LastName),
		)
	}
	fmt.Println(t.Render())
	return nil
}

func handlePayments(u *uiContext) error {
	u.header("Recent Payments")

	payments, err := u.client.GetPayments()
	t := u.newTable()
	t.AddColumn("Transaction ID", 18, "left")
	t.AddColumn("Date", 12, "left")
	t.AddColumn("Amount", 10, "right")
	t.AddColumn("Method", 10, "left")
	t.AddColumn("Status", 10, "left")

	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		t.AddStyledRow(statusStyle("SUCCESS").Style, "pay_123456789", "2026-05-30", "$45.67", "Stripe", "Success")
		t.AddStyledRow(statusStyle("SUCCESS").Style, "pay_123456790", "2024-01-14", "$123.45", "Stripe", "Success")
		t.AddStyledRow(statusStyle("FAILED").Style, "pay_123456791", "2024-01-13", "$67.89", "Credit", "Failed")
		fmt.Println(t.Render())
		return nil
	}

	for _, p := range payments {
		t.AddStyledRow(statusStyle(p.Status).Style,
			p.ID,
			p.CreatedAt.Format("2006-01-02"),
			fmt.Sprintf("$%.2f", p.Amount),
			p.Method,
			p.Status,
		)
	}
	fmt.Println(t.Render())
	return nil
}

func generateInvoice(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: Subscriber ID is required")
		u.muted("Usage: telecom-cli billing generate <subscriber_id>")
		return fmt.Errorf("missing subscriber id")
	}
	subscriberID := args[0]
	u.info("Generating invoice for subscriber: " + subscriberID)

	invoice, err := u.client.GenerateInvoice(subscriberID)
	t := u.newTable()
	t.AddColumn("Field", 14, "left")
	t.AddColumn("Value", 28, "left")

	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		u.success("Invoice generated successfully (simulated)")
		t.AddRow("Invoice #", "INV-000004")
		t.AddRow("Amount", "$45.67")
		t.AddRow("Due Date", "2024-02-15")
		fmt.Println(t.Render())
		return nil
	}

	u.success("Invoice generated successfully!")
	t.AddRow("Invoice #", invoice.ID)
	t.AddRow("Amount", fmt.Sprintf("$%.2f", invoice.Amount))
	t.AddRow("Due Date", invoice.DueDate.Format("2006-01-02"))
	t.AddRow("Status", invoice.Status)
	fmt.Println(t.Render())
	return nil
}
