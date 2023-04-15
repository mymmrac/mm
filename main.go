package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mymmrac/mm/utils"
)

const verboseFlag = "verbose"

func main() {
	rootCmd := &cobra.Command{
		Use:           "mm [flags] [expression]...",
		Short:         "mm is simple extendable expression evaluator",
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		ValidArgs:     []string{"expression\tExpression to evaluate"},
		Run: func(cmd *cobra.Command, args []string) {
			verbose, err := cmd.PersistentFlags().GetBool(verboseFlag)
			utils.Assert(err == nil, verboseFlag, "flag not found")

			debugger := &Debugger{}
			debugger.SetEnabled(verbose)

			if len(args) == 0 {
				runRepl(debugger)
			} else {
				// TODO: Support piping
				runImmediate(strings.Join(args, " "), debugger)
			}
		},
	}

	_ = rootCmd.PersistentFlags().BoolP(verboseFlag, "v", false, "Verbose output")

	utils.WalkCmd(rootCmd, utils.UpdateHelpFlag)
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "FATAL: %s\n", err)
		os.Exit(1)
	}
}

func runImmediate(expr string, debugger *Debugger) {
	executor := NewExecutor(debugger)

	result, err := executor.Execute(expr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if debugger.Enabled() {
		fmt.Println(debugger)
	}

	fmt.Println(result)
}

func runRepl(debugger *Debugger) {
	if _, err := tea.NewProgram(newModel(debugger)).Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "FATAL: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Bye!")
}
