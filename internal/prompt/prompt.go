package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// AskConfirm prompts the user for yes/no confirmation
func AskConfirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", message)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
		fmt.Fprintf(os.Stderr, "Sorry dude!, I don't know what to mean :?\n")
	}
}
