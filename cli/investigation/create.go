package investigation

import (
	"github.com/coinbase/dexter/cli/cliutil"
	"github.com/coinbase/dexter/engine"
	"github.com/coinbase/dexter/engine/helpers"
	"github.com/coinbase/dexter/facts"
	"github.com/coinbase/dexter/tasks"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"fmt"
	"os"
	"strconv"
)

var titleColor = color.New(color.FgHiGreen, color.Bold)

// A selectionWithArgs is a map from a task name to a set of arguments
type selectionWithArgs map[string][]string

// Map ordering is random, so to keep track of an ordered list
// of tasks, a map is created with integer keys.  The selectionWithArgs
// values will themselves only ever have one key, as each index
// should only describe one task, but structuring things this
// way makes some operations easier.
type orderedSelectionWithArgs map[int]selectionWithArgs

func createInvestigation(cmd *cobra.Command, args []string) {
	// Print an interesting welcome message
	titleColor.Println(`      _           _
     | |         | |
   __| | _____  _| |_ ___ _ __
  / _  |/ _ \ \/ / __/ _ \ '__|
 | (_| |  __/>  <| ||  __/ |
  \__,_|\___/_/\_\\__\___|_|

`)

	// Create a new investigation struct, interacting with the user where required for each field
	id := helpers.NewDexterID()
	investigation := engine.Investigation{
		ID:             id,
		TaskList:       collectTasks(),
		Scope:          collectFacts(id),
		KillContainers: cliutil.AskYesNo(color.HiCyanString("Terminate containers in scope after tasks complete?"), false),
		KillHost:       cliutil.AskYesNo(color.HiCyanString("Terminate hosts in scope after tasks compelte?"), false),
		RecipientNames: cliutil.SelectFromList(engine.LoadInvestigatorNames(), "Which investigators should be able to access this report?", true, true),
		Issuer:         engine.Signature{Name: engine.LocalInvestigatorName()},
	}

	// Sign the investigation, prompting the user to decrypt their key
	color.Yellow("The investigation will now be signed...")
	investigation.Sign(helpers.LoadLocalKey(cliutil.CollectPassword))

	// Upload the investigation to S3, reporting any errors
	err := investigation.Upload()
	if err != nil {
		color.HiRed("error uploading investigation: " + err.Error())
		os.Exit(1)
	} else {
		titleColor.Println("Investigation Uploaded!")
	}
}

// Drop into a command line loop to collect tasks to include in this investigation
func collectTasks() selectionWithArgs {
	color.HiCyan("Select tasks to run in this investigation, for more information try 'help'")
	numberedTasks := orderedSelectionWithArgs{}
	for {
		task, args := collectTask()
		switch task {
		case "exit":
			os.Exit(0)
		case "help":
			printTaskHelp()
		case "ls":
			listIncludedItems(numberedTasks)
		case "rm":
			removeItemsByNumber(numberedTasks, args)
		case "":
			if len(numberedTasks) == 0 {
				color.HiRed("please select at least one task")
				continue
			} else {
				return unorder(numberedTasks)
			}
		default:
			numberedTasks = addNewTask(task, args, numberedTasks)
		}
	}
	return unorder(numberedTasks)
}

// Drop into a command line loop to collect tasks to include in this investigation
func collectFacts(salt string) selectionWithArgs {
	color.HiCyan("Select facts to scope this investigation, for more information try 'help'")
	numberedFacts := orderedSelectionWithArgs{}
	for {
		task, args := collectFact()
		switch task {
		case "exit":
			os.Exit(0)
		case "help":
			printFactHelp()
		case "ls":
			listFacts(numberedFacts)
		case "rm":
			removeItemsByNumber(numberedFacts, args)
		default:
			if task == "" {
				return unorder(numberedFacts)
			}
			numberedFacts = addNewFact(task, args, numberedFacts, salt)
		}
	}
	return unorder(numberedFacts)
}

// Add a new fact to the current ordered list, printing any errors that make the selection invalid
func addNewFact(task string, args []string, numberedFacts orderedSelectionWithArgs, salt string) orderedSelectionWithArgs {
	if check, ok := facts.Facts[task]; ok {
		if len(args) < check.MinimumArguments {
			color.HiRed("not enough arguments, required: %d, provided: %d", check.MinimumArguments, len(args))
			return numberedFacts
		}
		if dedupFail(task, args, unorder(numberedFacts)) {
			color.HiRed("identical task and arguments already added")
			return numberedFacts
		}
		if check.Private {
			for i, arg := range args {
				args[i] = facts.Hash(arg, salt)
			}

		}
		numberedFacts[findSlot(numberedFacts)] = selectionWithArgs{
			task: args,
		}
		color.Green("ADDED: " + helpers.StringWithArgs(task, args, check.Private))
	} else {
		if task == "" {
			return numberedFacts
		}
		color.HiRed("unknown fact: %s", task)
	}
	return numberedFacts

}

// Add a new task to the current ordered list, printing any errors that make the selection invalid
func addNewTask(task string, args []string, numberedTasks orderedSelectionWithArgs) orderedSelectionWithArgs {
	if t, ok := tasks.Tasks[task]; ok {
		if len(args) < t.MinimumArguments {
			color.HiRed("not enough arguments, required: %d, provided: %d", t.MinimumArguments, len(args))
			return numberedTasks
		}
		if dedupFail(task, args, unorder(numberedTasks)) {
			color.HiRed("identical task and arguments already added")
			return numberedTasks
		}
		numberedTasks[findSlot(numberedTasks)] = selectionWithArgs{
			task: args,
		}
		color.Green("ADDED: " + helpers.StringWithArgs(task, args, false))
	} else {
		if task == "" {
			return numberedTasks
		}
		color.HiRed("unknown task: %s", task)
	}
	return numberedTasks

}

// List the tasks in order from lowest number to highest number
func listIncludedItems(numberedTasks orderedSelectionWithArgs) {
	keys := getOrderedKeys(numberedTasks)
	for _, num := range keys {
		t := numberedTasks[num]
		fmt.Print("[" + color.HiCyanString(strconv.Itoa(num)) + "]: ")
		for _, str := range helpers.TaskStrings(t) {
			color.HiYellow(str)
		}
	}
}

func listFacts(numberedTasks orderedSelectionWithArgs) {
	keys := getOrderedKeys(numberedTasks)
	for _, num := range keys {
		t := numberedTasks[num]
		fmt.Print("[" + color.HiCyanString(strconv.Itoa(num)) + "]: ")
		for k, v := range t { // there will only be one of these
			checker, ok := facts.Get(k)
			if !ok {
				color.HiRed("attempted to print fact that doesn't exist, should not be possible.  Different versions of Dexter?")
			} else {
				color.HiYellow(helpers.StringWithArgs(k, v, checker.Private))
			}
		}
	}
}

// Remove all the tasks specified by number
func removeItemsByNumber(numberedTasks orderedSelectionWithArgs, args []string) {
	for _, numstr := range args {
		num, err := strconv.Atoi(numstr)
		if err != nil {
			color.HiRed("bad selection, not an int")
			continue
		}
		if _, ok := numberedTasks[num]; !ok {
			color.HiRed("bad selection, task doesn't exist")
			continue
		}
		strs := helpers.TaskStrings(numberedTasks[num])
		if len(strs) != 1 {
			delete(numberedTasks, num)
			continue
		}
		color.Red("DELETED: " + strs[0])
		delete(numberedTasks, num)
	}
}

// Given the numbered tasks map, return the keys in numerical order
func getOrderedKeys(numberedTasks orderedSelectionWithArgs) []int {
	ordered := []int{}
	i := 0
	for len(ordered) != len(numberedTasks) {
		if _, ok := numberedTasks[i]; ok {
			ordered = append(ordered, i)
		}
		i += 1
	}
	return ordered
}

// Check if a new task name and args already exists in a set of tasks
func dedupFail(newTask string, newArgs []string, set selectionWithArgs) bool {
	for task, args := range set {
		if task == newTask && argsEqual(args, newArgs) {
			return true
		}
	}
	return false
}

// Compare two string slices, respecting order
func argsEqual(args []string, cmp []string) bool {
	if len(args) != len(cmp) {
		return false
	}
	for i, val := range args {
		if cmp[i] != val {
			return false
		}
	}
	return true
}

// Return the lowest value key available in the map
func findSlot(tasks orderedSelectionWithArgs) int {
	i := 0
	for {
		if _, ok := tasks[i]; !ok {
			return i
		}
		i += 1
	}
}

// Take a map from int to task, and return just the tasks
func unorder(tasks orderedSelectionWithArgs) selectionWithArgs {
	unordered := map[string][]string{}
	for _, task := range tasks {
		for k, v := range task {
			unordered[k] = v
		}
	}
	return unordered
}

// Print the command line and get user input back for task selection
func collectTask() (string, []string) {
	selections := []prompt.Suggest{}
	for k, v := range tasks.Tasks {
		if k == "example-task" {
			continue
		}
		selections = append(selections, prompt.Suggest{Text: k, Description: v.Description})
	}
	selections = append(selections, prompt.Suggest{Text: "help", Description: "print usage information for this promp"})
	completer := func(d prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(selections, d.GetWordBeforeCursor(), true)
	}
	input := cliutil.SplitArguments(prompt.Input("task [done] > ", completer))
	if len(input) < 1 {
		return "", []string{}
	}
	return input[0], input[1:]
}

// Print the command line and get user input back for task selection
func collectFact() (string, []string) {
	selections := []prompt.Suggest{}
	for k, v := range facts.Facts {
		if k == "example-fact" {
			continue
		}
		selections = append(selections, prompt.Suggest{Text: k, Description: v.Description})
	}
	selections = append(selections, prompt.Suggest{Text: "help", Description: "print usage information for this promp"})
	completer := func(d prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(selections, d.GetWordBeforeCursor(), true)
	}
	input := cliutil.SplitArguments(prompt.Input("fact [done] > ", completer))
	if len(input) < 1 {
		return "", []string{}
	}
	return input[0], input[1:]
}

// Print a help message for using the task selection CLI
func printTaskHelp() {
	promptColor := color.New(color.FgWhite)
	taskColor := color.New(color.FgHiYellow)
	argColor := color.New(color.FgGreen)
	color.White("\nDexter task selection:")
	color.White("\nStart typing a task name, and supported tasks will be autocompleted.")
	color.White("Add whitespace-separated arguments, if needed, like this:")
	promptColor.Print("\n\ttask [done] > ")
	taskColor.Print("my-task ")
	argColor.Println("arg1 arg2 arg3")
	color.White("\n'ls' will show all currently added tasks")
	color.White("'rm' will let you remove a task from this investigation\n")
	color.White("\nWhen you are done, enter an empty line to exit task selection")
	color.White("'exit' will cancel this investigation")
}

// Print a help message for using the fact selection CLI
func printFactHelp() {
	promptColor := color.New(color.FgWhite)
	taskColor := color.New(color.FgHiYellow)
	argColor := color.New(color.FgGreen)
	color.White("\nDexter fact selection:")
	color.White("\nStart typing a fact name, and supported facts will be autocompleted.")
	color.White("Add whitespace-separated arguments, if needed, like this:")
	promptColor.Print("\n\tfact [done] > ")
	taskColor.Print("my-task ")
	argColor.Println("arg1 arg2 arg3")
	color.White("\n'ls' will show all currently added facts")
	color.White("'rm' will let you remove a fact from this investigation\n")
	color.White("\nWhen you are done, enter an empty line to exit factc selection")
	color.White("'exit' will cancel this investigation")
}
