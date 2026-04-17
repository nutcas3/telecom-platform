package commands

import (
	"fmt"
)

func HandleBilling(args []string) {
	if len(args) == 0 {
		showBillingHelp()
		return
	}

	command := args[0]
	switch command {
	case "invoices":
		handleInvoices(args[1:])
	case "payments":
		handlePayments(args[1:])
	case "generate":
		generateInvoice(args[1:])
	default:
		fmt.Printf("Unknown billing command: %s\n", command)
		showBillingHelp()
	}
}

func showBillingHelp() {
	fmt.Println("Billing Management")
	fmt.Println("Usage: telecom-cli billing <command> [options]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  invoices                - List all invoices")
	fmt.Println("  payments                - List all payments")
	fmt.Println("  generate <subscriber>   - Generate invoice for subscriber")
}

func handleInvoices(args []string) {
	fmt.Println("Recent Invoices:")
	fmt.Println("Invoice #    Date        Status    Amount    Subscriber")
	fmt.Println("------------------------------------------------------------")
	fmt.Println("INV-000001   2024-01-15  Paid      $45.67    John Doe")
	fmt.Println("INV-000002   2024-01-15  Pending   $123.45   Jane Smith")
	fmt.Println("INV-000003   2024-01-14  Overdue   $67.89    Bob Johnson")
}

func handlePayments(args []string) {
	fmt.Println("Recent Payments:")
	fmt.Println("Transaction ID    Date        Amount    Method     Status")
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("pay_123456789    2024-01-15  $45.67    Stripe     Success")
	fmt.Println("pay_123456790    2024-01-14  $123.45   Stripe     Success")
	fmt.Println("pay_123456791    2024-01-13  $67.89    Credit     Failed")
}

func generateInvoice(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Subscriber ID is required")
		fmt.Println("Usage: telecom-cli billing generate <subscriber_id>")
		return
	}

	subscriberID := args[0]
	fmt.Printf("Generating invoice for subscriber: %s\n", subscriberID)
	fmt.Println("Invoice generated successfully!")
	fmt.Println("Invoice #: INV-000004")
	fmt.Println("Amount: $45.67")
	fmt.Println("Due Date: 2024-02-15")
}
