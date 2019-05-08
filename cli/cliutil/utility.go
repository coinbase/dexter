//
// The cliutil package contains helper functions specific to the command line interface.
//
package cliutil

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/coinbase/dexter/util"
)

//
// Read a string from the command line with a prompt.  If a "deefalt" string is provided,
// this will be provided as an option for the user to select automatically.  To prevent
// the user's input from echoing, the hidden argument should be set to true.  Users cannot
// provide an empty string or only whitespace, the prompt will loop until some input is
// provided.
//
func ReadString(prompt, deefalt string, hidden bool) string {
	var stdin = bufio.NewReader(os.Stdin)
	var text string
	for {
		if deefalt == "" {
			fmt.Print(fmt.Sprintf("%s > ", prompt))
		} else {
			fmt.Print(fmt.Sprintf("%s [%s] > ", prompt, deefalt))
		}
		var line string
		if hidden {
			line_bytes, _ := terminal.ReadPassword(syscall.Stdin)
			line = string(line_bytes)
			fmt.Print("\n")
		} else {
			line, _ = stdin.ReadString('\n')
		}
		text = strings.TrimSpace(line)
		if text == "" {
			if deefalt != "" {
				return deefalt
			}
		} else {
			break
		}
	}
	return text
}

//
// Ask a command line user to provide a new password, taking the password
// twice to confirm no errors in typing.
//
func CollectNewPassword() string {
	password := ""
	confirmed_password := false
	for !confirmed_password {
		password = ReadString("Set a new password", "", true)
		check := ReadString("Confirm", "", true)
		if password == check {
			confirmed_password = true
		} else {
			color.HiRed("Password mismatch, please try again")
		}
	}
	return password
}

//
// Retrieve a previously defined password from a user.
//
func CollectPassword() string {
	return ReadString("Password", "", true)
}

//
// Prompt the user for a yes or no question, with a default answer defined by the
// second argument.
//
func AskYesNo(str string, defaultTrue bool) bool {
	deefalt := ""
	if defaultTrue {
		deefalt = "y"
	} else {
		deefalt = "n"
	}

	for {
		answer := ReadString(str+" y/n", deefalt, false)
		if answer == "y" {
			return true
		} else if answer == "n" {
			return false
		} else {
			fmt.Println("\"y\" or \"n\", please")
		}
	}

	return false
}

//
// Prompt the user to selection options from a list, passing a list, a promp string, the default state
// (true = selected) for all members, and a boolean to indicate if at least one selection is required.
//
func SelectFromList(list []string, prompt string, defaultSelected, requiredResponse bool) []string {
	ordered := make(map[int]string)
	i := 0
	for _, item := range list {
		ordered[i] = item
		i += 1
	}

	selected := []string{}
	if defaultSelected {
		selected = list
	}

	color.HiCyan(prompt)
	for {
		printSelectionList(ordered, selected)
		change := ReadString("Choose number to toggle", "done", false)
		if change == "done" {
			if len(selected) > 0 {
				break
			} else {
				if !requiredResponse {
					break
				}
				color.Red("Please select a minimum of one")
			}
		} else {
			selected = toggleSelection(change, ordered, selected)
		}
	}
	return selected
}

func toggleSelection(change string, orderedOptions map[int]string, selectedOptions []string) []string {
	i, err := strconv.Atoi(change)
	if err != nil {
		return selectedOptions
	}
	option, ok := orderedOptions[i]
	if !ok {
		return selectedOptions
	}
	if util.StringsInclude(selectedOptions, option) {
		return util.StringsSubtract(selectedOptions, option)
	}
	return append(selectedOptions, option)
}

func printSelectionList(orderedOptions map[int]string, selectedOptions []string) {
	selected := color.New(color.FgHiYellow, color.Bold)
	for i := 0; i < len(orderedOptions); i++ {
		if util.StringsInclude(selectedOptions, orderedOptions[i]) {
			fmt.Println(fmt.Sprintf("%d. [", i) + selected.Sprint("*") + fmt.Sprintf("]\t%s", orderedOptions[i]))
		} else {
			fmt.Println(fmt.Sprintf("%d. [ ]\t%s", i, orderedOptions[i]))
		}
	}
}

//
// SplitArguments takes a string of arguments and splits them respecting quoted whitespace
//
func SplitArguments(args string) []string {
	if args == "" {
		return []string{}
	}
	r := csv.NewReader(strings.NewReader(args))
	r.Comma = ' '
	record, err := r.Read()
	if err != nil {
		color.HiRed("error parsing arguments: " + err.Error())
	}
	return record
}
