package pipeline

import (
	"strconv"

	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var listPipelinesCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lista los pipelines del proyecto con su último resultado",
	Example: `  azdevops pipelines list
  azdevops pipelines list --filter "deploy-*" --folder "\\infra"`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		filter, _ := c.Flags().GetString("filter")
		folder, _ := c.Flags().GetString("folder")
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		var defs []models.BuildDefinition
		if err := ui.Spinner("Cargando pipelines...", func() (err error) {
			defs, err = pipeline.ListDefinitions(client, filter, folder)
			return err
		}); err != nil {
			return err
		}
		if out == "json" {
			if defs == nil {
				defs = []models.BuildDefinition{}
			}
			return ui.PrintJSON(defs)
		}
		if len(defs) == 0 {
			ui.Info("No se encontraron pipelines.")
			return nil
		}
		rows := make([][]string, 0, len(defs))
		for _, d := range defs {
			id := ""
			if d.ID != nil {
				id = strconv.FormatInt(*d.ID, 10)
			}
			last, when, branch := "-", "-", "-"
			if b := d.LatestBuild; b != nil {
				last, when, branch = buildState(b), ui.FormatTime(b.QueueTime), cmd.ShortBranch(b.SourceBranch)
			}
			rows = append(rows, []string{id, cmd.Str(d.Name), cmd.Str(d.Path), last, branch, when})
		}
		ui.Table([]string{"id", "nombre", "carpeta", "última ejecución", "rama", "fecha"}, rows)
		return nil
	},
}

func init() {
	listPipelinesCmd.Flags().StringP("filter", "f", "", "Filtro por nombre, admite comodines")
	listPipelinesCmd.Flags().String("folder", "", "Carpeta (p. ej. \\infra)")
	cmd.Pipelines.AddCommand(listPipelinesCmd)
}
