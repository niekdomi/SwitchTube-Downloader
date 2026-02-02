// Package table provides functions to display information in markdown tables.
package table

import (
	"os"
	"strconv"

	"switchtube-downloader/internal/helper/ui/ansi"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

const createAccessTokenURL = "https://tube.switch.ch/access_tokens"

// DisplayInstructions shows token creation instructions in a table.
func DisplayInstructions() {
	data := [][]string{
		{ansi.Bold + "1." + ansi.Reset + " Visit: " + createAccessTokenURL},
		{ansi.Bold + "2." + ansi.Reset + " Click 'Create New Token'"},
		{ansi.Bold + "3." + ansi.Reset + " Copy the generated token"},
		{ansi.Bold + "4." + ansi.Reset + " Paste it below"},
	}

	table := createTable()
	table.Header("Token Creation Instructions")
	table.Bulk(data)
	table.Render()
}

// DisplayTokenInfo shows token information in a table.
func DisplayTokenInfo(service string, username string, status string, maskedToken string, tokenLength int) {
	data := [][]string{
		{"Service", service},
		{"User", username},
		{"Token", maskedToken},
		{"Length", strconv.Itoa(tokenLength) + " characters"},
		{"Status", status},
	}

	table := createTable()
	table.Header("Token Information")
	table.Bulk(data)
	table.Render()
}

// createTable creates a markdown table.
func createTable() *tablewriter.Table {
	cfg := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{AutoFormat: tw.Off},
		},
	}

	return tablewriter.NewTable(os.Stdout,
		tablewriter.WithSymbols(tw.NewSymbols(tw.StyleRounded)),
		tablewriter.WithConfig(cfg),
	)
}
