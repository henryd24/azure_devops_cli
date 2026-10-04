package pipeline

import (
	"fmt"

	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var getPipelineCmd = &cobra.Command{
	Use:   "get",
	Short: "Obtiene la definición de uno o más pipelines (por nombre o ID)",
	Example: `  azdevops pipelines get --name "MiPipeline*"
  azdevops pipelines get --id 123`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		id, _ := c.Flags().GetInt("id")
		name, _ := c.Flags().GetString("name")
		if id == 0 && name == "" {
			if !ui.Interactive() {
				return fmt.Errorf("debes proporcionar --name o --id")
			}
			if id, err = cmd.PickPipeline(client, "Pipeline"); err != nil {
				return err
			}
		}
		if id != 0 {
			def, err := pipeline.GetBuildDefinitionByID(client, id)
			if err != nil {
				return err
			}
			return ui.PrintJSON(def)
		}
		defs, err := pipeline.GetBuildDefinitionByName(client, name)
		if err != nil {
			return err
		}
		if len(defs) == 0 {
			return fmt.Errorf("no se encontró ningún pipeline con el nombre '%s'", name)
		}
		return ui.PrintJSON(defs)
	},
}

var getPipelineByIDCmd = &cobra.Command{
	Use:        "get-by-id",
	Short:      "Obtiene un pipeline por su ID",
	Deprecated: "usa 'pipelines get --id'",
	RunE: func(c *cobra.Command, args []string) error {
		return getPipelineCmd.RunE(c, args)
	},
}

func init() {
	getPipelineCmd.Flags().StringP("name", "n", "", "Nombre del pipeline (admite comodines)")
	getPipelineCmd.Flags().IntP("id", "i", 0, "ID del pipeline")
	_ = getPipelineCmd.RegisterFlagCompletionFunc("id", cmd.CompletePipelineIDs)
	getPipelineByIDCmd.Flags().IntP("id", "i", 0, "ID del pipeline")
	getPipelineByIDCmd.Flags().StringP("name", "n", "", "")
	_ = getPipelineByIDCmd.Flags().MarkHidden("name")
	cmd.Pipelines.AddCommand(getPipelineCmd, getPipelineByIDCmd)
}
