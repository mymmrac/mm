package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mymmrac/mm/debugger"
	"github.com/mymmrac/mm/executor/v2"
	"github.com/mymmrac/mm/repl"
	"github.com/mymmrac/mm/utils"
)

const (
	verboseFlag   = "verbose"
	precisionFlag = "precision"
)

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

			precision, err := cmd.PersistentFlags().GetInt32(precisionFlag)
			utils.Assert(err == nil, precisionFlag, "flag not found")

			debug := &debugger.Debugger{}
			debug.SetEnabled(verbose)

			fi, err := os.Stdin.Stat()
			isPiped := err == nil && (fi.Mode()&os.ModeNamedPipe) != 0

			if isPiped {
				expr, readErr := io.ReadAll(os.Stdin)
				utils.Assert(readErr == nil, "reading from stdin:", readErr)
				runImmediate(string(expr), precision, debug)
			} else if len(args) != 0 {
				runImmediate(strings.Join(args, " "), precision, debug)
			} else {
				runRepl(precision, debug)
			}
		},
	}

	_ = rootCmd.PersistentFlags().BoolP(verboseFlag, "v", false, "Verbose output")
	_ = rootCmd.PersistentFlags().Int32P(precisionFlag, "p", 16, "Precision")

	utils.WalkCmd(rootCmd, utils.UpdateHelpFlag)
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "FATAL: %s\n", err)
		os.Exit(1)
	}
}

func runImmediate(expr string, precision int32, debugger *debugger.Debugger) {
	exec := executor.NewExecutor(debugger)

	result, err := exec.Execute(expr, precision)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if debugger.Enabled() {
		fmt.Println(debugger)
	}

	fmt.Println(result)
}

func runRepl(precision int32, debugger *debugger.Debugger) {
	if _, err := tea.NewProgram(repl.NewModel(debugger, precision)).Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "FATAL: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Bye!")
}
