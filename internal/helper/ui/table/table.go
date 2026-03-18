// Package table provides functions to display information in formatted tables.
package table

import (
	"fmt"

	"switchtube-downloader/internal/helper/ui/styles"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

const createAccessTokenURL = "https://tube.switch.ch/access_tokens"

var (
	borderStyle = lipgloss.NewStyle().Foreground(styles.Cyan)
	headerStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	cellStyle   = lipgloss.NewStyle().Padding(0, 1)
	keyStyle    = cellStyle.Foreground(styles.Cyan)
)

func newTable() *table.Table {
	return table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		BorderColumn(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			if col == 0 {
				return keyStyle
			}

			return cellStyle
		})
}

// DisplayInstructions shows token creation instructions in a table.
func DisplayInstructions() {
	t := newTable().
		Headers("Token creation instructions").
		Row("1. Visit: " + createAccessTokenURL).
		Row("2. Click 'Create New Token'").
		Row("3. Copy the generated token").
		Row("4. Paste it below")

	fmt.Println(t.Render())
}

// DisplayTokenInfo shows token information in a table.
func DisplayTokenInfo(service string, username string, valid bool, maskedToken string, tokenLength int) {
	var status string
	if valid {
		status = styles.Success.Render("Valid")
	} else {
		status = styles.Error.Render("Invalid")
	}

	t := newTable().
		Headers("Field", "Value").
		Row("Service", service).
		Row("User", username).
		Row("Token", maskedToken).
		Row("Length", fmt.Sprintf("%d characters", tokenLength)).
		Row("Status", status)

	fmt.Println(t.Render())
}
