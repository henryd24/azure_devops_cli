package pipeline

import (
	"fmt"

	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var updatePipelineCmd = &cobra.Command{
	Use:   "update",
	Short: "Actualiza nombre, YAML, repositorio, rama o conexión de un pipeline",
	Example: `  azdevops pipelines update --id 123 --new-name NuevoNombre --yaml-path ci.yml
  azdevops pipelines update   # modo guiado`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		id, err := resolvePipelineID(c, client)
		if err != nil {
			return err
		}
		f := c.Flags()
		opts := pipeline.UpdateOptions{}
		opts.NewName, _ = f.GetString("new-name")
		opts.YAMLPath, _ = f.GetString("yaml-path")
		opts.RepoName, _ = f.GetString("repo-name")
		opts.Branch, _ = f.GetString("branch")
		opts.ServiceConnectionID, _ = f.GetString("service-connection")

		if opts.IsEmpty() {
			if !ui.Interactive() {
				return fmt.Errorf("debes proporcionar al menos un valor para actualizar")
			}
			current, err := pipeline.GetBuildDefinitionByID(client, id)
			if err != nil {
				return err
			}
			curName, curYAML, curBranch := cmd.Str(current.Name), "", ""
			if current.Process != nil {
				curYAML = cmd.Str(current.Process.YAMLFilename)
			}
			if current.Repository != nil {
				curBranch = cmd.ShortBranch(cmd.Str(current.Repository.DefaultBranch))
			}
			ui.Info("Deja el valor actual para no modificarlo")
			if v, err := ui.Input("Nombre", "", curName, ui.Required); err != nil {
				return err
			} else if v != curName {
				opts.NewName = v
			}
			if v, err := ui.Input("Ruta del YAML", "", curYAML, ui.Required); err != nil {
				return err
			} else if v != curYAML {
				opts.YAMLPath = v
			}
			if v, err := ui.Input("Rama por defecto", "", curBranch, ui.Required); err != nil {
				return err
			} else if v != curBranch {
				opts.Branch = v
			}
			if opts.IsEmpty() {
				ui.Info("No hay cambios.")
				return nil
			}
		}

		updated, err := pipeline.UpdateBuildDefinition(client, id, opts)
		if err != nil {
			return err
		}
		ui.Success("Pipeline '%s' (ID: %d) actualizado", cmd.Str(updated.Name), id)
		if cmd.WantsJSON() {
			return ui.PrintJSON(updated)
		}
		return nil
	},
}

func init() {
	addPipelineFlags(updatePipelineCmd, "a actualizar")
	f := updatePipelineCmd.Flags()
	f.String("new-name", "", "Nuevo nombre del pipeline")
	f.String("yaml-path", "", "Nueva ruta al archivo YAML")
	f.String("repo-name", "", "Nuevo repositorio en formato 'org/repo'")
	f.String("branch", "", "Nueva rama por defecto")
	f.String("service-connection", "", "Nuevo ID de la conexión de servicio")
	cmd.Pipelines.AddCommand(updatePipelineCmd)
}
