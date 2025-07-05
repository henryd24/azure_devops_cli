package cmd

import (
	"github.com/spf13/cobra"
)

var Variables = &cobra.Command{
	Use:   "variables",
	Short: "Comandos para trabajar con variables",
	Long:  "Comandos para trabajar con variables en Azure DevOps, como Variable Groups.",
}

var Pipelines = &cobra.Command{
	Use:   "pipelines",
	Short: "Comandos para trabajar con pipelines",
	Long:  "Comandos para trabajar con pipelines en Azure DevOps, incluyendo la creación, actualización y eliminación de pipelines.",
}

func init() {
	RootCmd.AddCommand(Variables)
	RootCmd.AddCommand(Pipelines)
}
