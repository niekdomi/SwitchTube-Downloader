package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Input prompts the user for input and returns the entered string.
func Input(prompt string) string {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	return strings.TrimSpace(input)
}

// Confirm prompts the user for a yes/no confirmation and returns true for yes.
func Confirm(format string, args ...any) bool {
	prompt := fmt.Sprintf(format, args...)
	response := Input(prompt + " (y/N): ")
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes"
}
