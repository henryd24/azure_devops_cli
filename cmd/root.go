package cmd

import (
	"errors"
	"os"

	"azuredevops/azdevops"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var globalFlags struct {
	org     string
	project string
	pat     string
	profile string
	output  string
	noInput bool
	debug   bool
}

// InteractiveRun lo registra el paquete cmd/interactive para abrir el menú
// cuando se ejecuta "azdevops" sin argumentos en una terminal.
var InteractiveRun func(cmd *cobra.Command) error

var RootCmd = &cobra.Command{
	Use:   "azdevops",
	Short: "CLI para Azure DevOps",
	Long: `CLI no oficial para Azure DevOps: Variable Groups, Pipelines y Seguridad.

Ejecuta "azdevops" sin argumentos en una terminal para abrir el modo interactivo,
o "azdevops login" para guardar tu organización, proyecto y PAT.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		allowed := !globalFlags.noInput && os.Getenv("AZDEVOPS_NO_INPUT") == "" && os.Getenv("CI") == ""
		ui.SetInteractive(allowed)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if ui.Interactive() && InteractiveRun != nil {
			return InteractiveRun(cmd)
		}
		return cmd.Help()
	},
}

func init() {
	RootCmd.Version = azdevops.Version
	pf := RootCmd.PersistentFlags()
	pf.StringVar(&globalFlags.org, "org", "", "Organización de Azure DevOps (env: AZURE_ORG)")
	pf.StringVar(&globalFlags.project, "project", "", "Proyecto de Azure DevOps (env: AZURE_PROJECT)")
	pf.StringVar(&globalFlags.pat, "pat", "", "Personal Access Token (env: AZURE_PAT). Preferible usar la variable de entorno o 'azdevops login'")
	pf.StringVar(&globalFlags.profile, "profile", "", "Perfil guardado a usar (env: AZDEVOPS_PROFILE)")
	pf.StringVarP(&globalFlags.output, "output", "o", "", "Formato de salida: json o table")
	pf.BoolVar(&globalFlags.noInput, "no-input", false, "Nunca preguntar de forma interactiva (env: AZDEVOPS_NO_INPUT)")
	pf.BoolVar(&globalFlags.debug, "debug", false, "Muestra las peticiones HTTP realizadas")

	_ = RootCmd.RegisterFlagCompletionFunc("output", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "table"}, cobra.ShellCompDirectiveNoFileComp
	})
	RootCmd.SetVersionTemplate("azdevops {{.Version}}\n")
}

// Execute ejecuta la CLI y termina el proceso con el código adecuado.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		if errors.Is(err, ui.ErrAborted) {
			ui.Warn("%v", err)
			os.Exit(130)
		}
		ui.Errorf("%v", err)
		os.Exit(1)
	}
}
