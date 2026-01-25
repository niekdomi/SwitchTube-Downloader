// Package table provides functions to display information in markdownmarkdown .
package table

import (
	"os"
	"strconv"

	"switchtube-downloader/internal/helper/ui/colors"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

const createAccessTokenURL = "https://tube.switch.ch/access_tokens"

// DisplayInstructions shows token creation instructions in a rounded table.
func DisplayInstructions() {
	data := [][]string{
		{colors.Bold + "1." + colors.Reset + " Visit: " + createAccessTokenURL},
		{colors.Bold + "2." + colors.Reset + " Click 'Create New Token'"},
		{colors.Bold + "3." + colors.Reset + " Copy the generated token"},
		{colors.Bold + "4." + colors.Reset + " Paste it below"},
	}

	table := createRoundedTable()
	table.Header("Token Creation Instructions")
	table.Bulk(data)
	table.Render()
}

// DisplayTokenInfo shows token information in a rounded table.
func DisplayTokenInfo(service string, username string, status string, maskedToken string, tokenLength int) {
	data := [][]string{
		{"Service", service},
		{"User", username},
		{"Token", maskedToken},
		{"Length", strconv.Itoa(tokenLength) + " characters"},
		{"Status", status},
	}

	table := createRoundedTable()
	table.Header("Token Information")
	table.Bulk(data)
	table.Render()
}

// createRoundedTable creates a table with rounded corners and preserved casing.
func createRoundedTable() *tablewriter.Table {
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
