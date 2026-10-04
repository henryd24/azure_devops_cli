package pipeline

import (
	"fmt"

	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var deletePipelineCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Elimina un pipeline y sus retenciones",
	Example: `  azdevops pipelines delete --id 123 --yes`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		id, err := resolvePipelineID(c, client)
		if err != nil {
			return err
		}
		def, err := pipeline.GetBuildDefinitionByID(client, id)
		if err != nil {
			return err
		}
		yes, _ := c.Flags().GetBool("yes")
		if err := cmd.Confirm(fmt.Sprintf("¿Eliminar el pipeline '%s' (ID %d) y todas sus retenciones?", cmd.Str(def.Name), id), yes); err != nil {
			return err
		}
		leases, err := pipeline.DeletePipeline(client, id)
		if leases > 0 {
			ui.Success("%d retención(es) eliminada(s)", leases)
		}
		if err != nil {
			return err
		}
		ui.Success("Pipeline '%s' (ID %d) eliminado", cmd.Str(def.Name), id)
		return nil
	},
}

func init() {
	addPipelineFlags(deletePipelineCmd, "a eliminar")
	deletePipelineCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
	cmd.Pipelines.AddCommand(deletePipelineCmd)
}
