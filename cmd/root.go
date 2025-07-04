package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{ // <- Ahora es exportada
	Use:   "azurevgcli",
	Short: "CLI para Azure DevOps",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
