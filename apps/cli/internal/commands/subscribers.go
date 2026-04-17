package commands

import (
	"fmt"

	"github.com/nutcas3/telecom-platform/apps/cli/internal/types"
)

// HandleSubscribers is the entry point for subscriber commands.
func HandleSubscribers(args []string, config *types.CLIConfig) error {
	u := newUIContext(config)
	if len(args) == 0 {
		showSubscribersHelp(u)
		return nil
	}

	u.connectivityBanner()

	command := args[0]
	switch command {
	case "list":
		return listSubscribers(u)
	case "create":
		return createSubscriber(u, args[1:])
	case "delete":
		return deleteSubscriber(u, args[1:])
	case "show":
		return showSubscriber(u, args[1:])
	default:
		u.errorln("Unknown subscribers command: " + command)
		showSubscribersHelp(u)
		return fmt.Errorf("unknown command: %s", command)
	}
}

func showSubscribersHelp(u *uiContext) {
	u.header("Subscriber Management")
	u.muted("Usage: telecom-cli subscribers <command> [options]")
	fmt.Println()

	t := u.newTable()
	t.AddColumn("Command", 24, "left")
	t.AddColumn("Description", 40, "left")
	t.AddRow("list", "List all subscribers")
	t.AddRow("show <imsi>", "Show subscriber details")
	t.AddRow("create <imsi> <name>", "Create a new subscriber")
	t.AddRow("delete <imsi>", "Delete a subscriber")
	fmt.Println(t.Render())
}

func listSubscribers(u *uiContext) error {
	u.header("Active Subscribers")

	subs, err := u.client.ListSubscribers()
	if err != nil || len(subs) == 0 {
		if err != nil {
			u.warn("Using placeholder data: " + err.Error())
		} else {
			u.muted("(API returned no subscribers - showing sample data)")
		}
		t := u.newTable()
		t.AddColumn("IMSI", 16, "left")
		t.AddColumn("Name", 22, "left")
		t.AddColumn("Status", 10, "left")
		t.AddColumn("Balance", 10, "right")
		t.AddStyledRow(statusStyle("ACTIVE").Style, "310260123456789", "John Doe", "Active", "$45.67")
		t.AddStyledRow(statusStyle("ACTIVE").Style, "310260123456790", "Jane Smith", "Active", "$123.45")
		t.AddStyledRow(statusStyle("INACTIVE").Style, "310260123456791", "Bob Johnson", "Inactive", "$0.00")
		fmt.Println(t.Render())
		return nil
	}

	t := u.newTable()
	t.AddColumn("IMSI", 16, "left")
	t.AddColumn("Name", 22, "left")
	t.AddColumn("Status", 10, "left")
	t.AddColumn("Balance", 10, "right")
	for _, s := range subs {
		t.AddStyledRow(statusStyle(s.Status).Style,
			s.IMSI, s.Name, s.Status, fmt.Sprintf("$%.2f", s.Balance))
	}
	fmt.Println(t.Render())
	return nil
}

func createSubscriber(u *uiContext, args []string) error {
	if len(args) < 2 {
		u.errorln("Error: IMSI and name are required")
		u.muted("Usage: telecom-cli subscribers create <imsi> <name>")
		return fmt.Errorf("missing arguments")
	}
	imsi, name := args[0], args[1]
	u.info(fmt.Sprintf("Creating subscriber: IMSI=%s, Name=%s", imsi, name))
	sub, err := u.client.CreateSubscriber(imsi, name)
	if err != nil {
		u.warn("API error, simulated success: " + err.Error())
		u.success("Subscriber created successfully (simulated)")
		return nil
	}
	u.success("Subscriber created successfully!")
	t := u.newTable()
	t.AddColumn("Field", 12, "left")
	t.AddColumn("Value", 24, "left")
	t.AddRow("IMSI", sub.IMSI)
	t.AddRow("Name", sub.Name)
	t.AddRow("Status", sub.Status)
	fmt.Println(t.Render())
	return nil
}

func deleteSubscriber(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: IMSI is required")
		u.muted("Usage: telecom-cli subscribers delete <imsi>")
		return fmt.Errorf("missing imsi")
	}
	imsi := args[0]
	u.info("Deleting subscriber: IMSI=" + imsi)
	if err := u.client.DeleteSubscriber(imsi); err != nil {
		u.warn("API error, simulated success: " + err.Error())
		u.success("Subscriber deleted successfully (simulated)")
		return nil
	}
	u.success("Subscriber deleted successfully!")
	return nil
}

func showSubscriber(u *uiContext, args []string) error {
	if len(args) < 1 {
		u.errorln("Error: IMSI is required")
		u.muted("Usage: telecom-cli subscribers show <imsi>")
		return fmt.Errorf("missing imsi")
	}
	imsi := args[0]
	u.header("Subscriber Details: " + imsi)

	sub, err := u.client.GetSubscriber(imsi)
	t := u.newTable()
	t.AddColumn("Field", 14, "left")
	t.AddColumn("Value", 28, "left")
	if err != nil {
		u.warn("Using placeholder data: " + err.Error())
		t.AddRow("IMSI", imsi)
		t.AddRow("Name", "John Doe")
		t.AddRow("Status", "Active")
		t.AddRow("Balance", "$45.67")
		t.AddRow("Data Used", "2.3GB / 10GB")
		t.AddRow("Voice Used", "45min / 500min")
		t.AddRow("SMS Used", "23 / 100")
		fmt.Println(t.Render())
		return nil
	}
	t.AddRow("IMSI", sub.IMSI)
	t.AddRow("Name", sub.Name)
	t.AddRow("Status", sub.Status)
	t.AddRow("Balance", fmt.Sprintf("$%.2f", sub.Balance))
	fmt.Println(t.Render())
	return nil
}
