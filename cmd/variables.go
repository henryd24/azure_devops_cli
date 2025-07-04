package cmd

import (
	"github.com/spf13/cobra"
)

var groupName string

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Comandos para obtener recursos",
	Long:  "Comandos para obtener recursos de Azure DevOps, como Variable Groups.",
}

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Comandos para actualizar recursos",
	Long:  "Comandos para actualizar recursos de Azure DevOps, como Variable Groups.",
}

func init() {
	RootCmd.AddCommand(GetCmd)
	RootCmd.AddCommand(UpdateCmd)
}
