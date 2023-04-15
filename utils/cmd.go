package utils

import "github.com/spf13/cobra"

func WalkCmd(cmd *cobra.Command, f func(*cobra.Command)) {
	f(cmd)

	for _, childCmd := range cmd.Commands() {
		WalkCmd(childCmd, f)
	}
}

func UpdateHelpFlag(cmd *cobra.Command) {
	cmd.InitDefaultHelpFlag()

	f := cmd.Flags().Lookup("help")
	if f != nil {
		if cmd.Name() != "" {
			f.Usage = "Help for " + cmd.Name()
		} else {
			f.Usage = "Help for this command"
		}
	}
}
