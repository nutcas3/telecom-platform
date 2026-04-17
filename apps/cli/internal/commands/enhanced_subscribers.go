package commands

import (
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// HandleSubscribersEnhanced delegates to the UI + API-connected subscribers handler.
// It also supports a few extra subcommands (activate/deactivate/balance/usage/search)
// that fall back to informative UI-styled placeholders until their API endpoints
// are wired up.
func HandleSubscribersEnhanced(args []string, config *types.CLIConfig) error {
	if len(args) == 0 {
		return HandleSubscribers(args, config)
	}

	command := args[0]
	switch command {
	case "list", "show", "create", "delete":
		return HandleSubscribers(args, config)
	case "update":
		return updateSubscriberUI(args[1:], config)
	case "activate", "deactivate":
		return toggleSubscriberUI(command, args[1:], config)
	case "balance":
		return balanceSubscriberUI(args[1:], config)
	case "usage":
		return usageSubscriberUI(args[1:], config)
	case "search":
		return searchSubscribersUI(args[1:], config)
	default:
		u := newUIContext(config)
		u.errorln("Unknown subscribers command: " + command)
		_ = HandleSubscribers(nil, config)
		return fmt.Errorf("unknown subscribers command: %s", command)
	}
}

func updateSubscriberUI(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) < 1 {
		u.errorln("Error: IMSI is required")
		u.muted("Usage: telecom-cli subscribers update <imsi>")
		return fmt.Errorf("missing imsi")
	}
	u.info("Update subscriber: " + args[0])
	u.muted("Update API endpoint is not yet wired; showing placeholder confirmation.")
	u.success("Subscriber update accepted (simulated)")
	return nil
}

func toggleSubscriberUI(cmd string, args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) < 1 {
		u.errorln("Error: IMSI is required")
		u.muted("Usage: telecom-cli subscribers " + cmd + " <imsi>")
		return fmt.Errorf("missing imsi")
	}
	imsi := args[0]
	u.info(cmd + " subscriber: " + imsi)
	u.muted("Activation API endpoint is not yet wired; showing placeholder confirmation.")
	u.success("Subscriber " + cmd + " completed (simulated)")
	return nil
}

func balanceSubscriberUI(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) < 1 {
		u.errorln("Error: IMSI is required")
		u.muted("Usage: telecom-cli subscribers balance <imsi>")
		return fmt.Errorf("missing imsi")
	}
	imsi := args[0]
	u.header("Subscriber Balance: " + imsi)

	sub, err := u.client.GetSubscriber(imsi)
	t := u.newTable()
	t.AddColumn("Field", 14, "left")
	t.AddColumn("Value", 24, "left")
	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		t.AddRow("IMSI", imsi)
		t.AddRow("Balance", "$45.67")
		t.AddRow("Currency", "USD")
		fmt.Println(t.Render())
		return nil
	}
	t.AddRow("IMSI", sub.IMSI)
	t.AddRow("Name", sub.Name)
	t.AddRow("Balance", fmt.Sprintf("$%.2f", sub.Balance))
	fmt.Println(t.Render())
	return nil
}

func usageSubscriberUI(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) < 1 {
		u.errorln("Error: IMSI is required")
		u.muted("Usage: telecom-cli subscribers usage <imsi>")
		return fmt.Errorf("missing imsi")
	}
	imsi := args[0]
	u.header("Subscriber Usage: " + imsi)
	u.muted("(Usage API endpoint not yet wired - showing placeholder)")

	t := u.newTable()
	t.AddColumn("Metric", 14, "left")
	t.AddColumn("Used", 14, "right")
	t.AddColumn("Limit", 14, "right")
	t.AddRow("Data", "2.3 GB", "10 GB")
	t.AddRow("Voice", "45 min", "500 min")
	t.AddRow("SMS", "23", "100")
	fmt.Println(t.Render())
	return nil
}

func searchSubscribersUI(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) < 1 {
		u.errorln("Error: search query is required")
		u.muted("Usage: telecom-cli subscribers search <query>")
		return fmt.Errorf("missing query")
	}
	u.header("Search Subscribers: " + args[0])

	subs, err := u.client.ListSubscribers()
	t := u.newTable()
	t.AddColumn("IMSI", 16, "left")
	t.AddColumn("Name", 22, "left")
	t.AddColumn("Status", 10, "left")
	t.AddColumn("Balance", 10, "right")

	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		t.AddStyledRow(statusStyle("ACTIVE").Style, "310260123456789", "John Doe", "Active", "$45.67")
		fmt.Println(t.Render())
		return nil
	}
	for _, s := range subs {
		t.AddStyledRow(statusStyle(s.Status).Style,
			s.IMSI, s.Name, s.Status, fmt.Sprintf("$%.2f", s.Balance))
	}
	fmt.Println(t.Render())
	return nil
}
