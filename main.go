package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/coinbase/dexter/cli/daemon"
	"github.com/coinbase/dexter/cli/investigation"
	"github.com/coinbase/dexter/cli/investigator"
	"github.com/coinbase/dexter/cli/report"
	"github.com/coinbase/dexter/engine/helpers"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	// Override aws keys with dexter-specific ones where available
	dexterAWSAccessKey := os.Getenv("DEXTER_AWS_ACCESS_KEY_ID")
	if dexterAWSAccessKey != "" {
		os.Setenv("AWS_ACCESS_KEY_ID", dexterAWSAccessKey)
	}
	dexterAWSSecretKey := os.Getenv("DEXTER_AWS_SECRET_ACCESS_KEY")
	if dexterAWSSecretKey != "" {
		os.Setenv("AWS_SECRET_ACCESS_KEY", dexterAWSSecretKey)
	}
	dexterAWSRegion := os.Getenv("DEXTER_AWS_REGION")
	if dexterAWSRegion != "" {
		os.Setenv("AWS_REGION", dexterAWSRegion)
	}

	handlePasswordInterrupts()
	var rootCmd = &cobra.Command{
		Use:   "dexter",
		Short: "Your friendly forensics expert",
		Long: `Dexter: Your Friendly Forensics Expert

Dexter is a tool build to facilitate the secure exectuion
and reporting of forensics tasks on remote hosts.  This
program contains functionallity to run a Dexter daemon
as well as interact with Dexter from the command line as an
investigator.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			helpers.BuildDemoPath()
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "docs",
		Short: "Update the docs directory",
		Long:  `This command generates markdown documentation in doc/`,
		Args:  cobra.MaximumNArgs(0),
		Run:   func(_ *cobra.Command, _ []string) { doc.GenMarkdownTree(rootCmd, "doc") },
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "bash",
		Short: "Generate a bash completion script",
		Long:  `This command generates a shell script for bash completion`,
		Args:  cobra.MaximumNArgs(0),
		Run:   func(_ *cobra.Command, _ []string) { rootCmd.GenBashCompletionFile("dexter.sh") },
	})

	rootCmd.AddCommand(daemon.CommandSuite())
	rootCmd.AddCommand(investigator.CommandSuite())
	rootCmd.AddCommand(investigation.CommandSuite())
	rootCmd.AddCommand(report.CommandSuite())

	rootCmd.PersistentFlags().StringVar(&helpers.LocalDemoPath, "demo", "", "run fom a local path for demo purposes, not S3")

	rootCmd.Execute()
}

//
// If a Ctrl+C is given while at a password prompt, the `stty -echo` that was used
// to prevent password echo will break the terminal.  Make sure to clean that up
// here for the best user experience.
//
func handlePasswordInterrupts() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			attrs := syscall.ProcAttr{
				Dir:   "",
				Env:   []string{},
				Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
				Sys:   nil}
			var ws syscall.WaitStatus
			pid, err := syscall.ForkExec(
				"/bin/stty",
				[]string{"stty", "echo"},
				&attrs)
			if err != nil {
				panic(err)
			}
			_, err = syscall.Wait4(pid, &ws, 0, nil)
			if err != nil {
				panic(err)
			}
			os.Exit(0)
		}
	}()
}
