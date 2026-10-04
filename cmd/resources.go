package cmd

import (
	"github.com/spf13/cobra"
)

var Variables = &cobra.Command{
	Use:     "variables",
	Aliases: []string{"vg", "variable-groups"},
	Short:   "Gestiona Variable Groups",
	Long:    "Comandos para trabajar con variables en Azure DevOps, como Variable Groups.",
}

var Pipelines = &cobra.Command{
	Use:     "pipelines",
	Aliases: []string{"pipeline", "pl"},
	Short:   "Gestiona y ejecuta pipelines",
	Long:    "Comandos para trabajar con pipelines en Azure DevOps: crear, actualizar, eliminar, ejecutar y seguir sus ejecuciones.",
}

var Security = &cobra.Command{
	Use:     "security",
	Aliases: []string{"sec"},
	Short:   "Gestiona grupos de seguridad y membresías",
}

func init() {
	RootCmd.AddCommand(Variables)
	RootCmd.AddCommand(Pipelines)
	RootCmd.AddCommand(Security)
}
