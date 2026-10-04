package variable_group

import (
	"fmt"

	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var createVariableGroupCmd = &cobra.Command{
	Use:   "create",
	Short: "Crea un nuevo Variable Group",
	Example: `  azdevops variables create --name MiGrupo -d "Descripción" -v clave1=valor1 -v secret:token=abc
  azdevops variables create   # modo guiado`,
	RunE: func(c *cobra.Command, args []string) error {
		name, err := cmd.ResolveStringFlag(c, "name", func() (string, error) {
			return ui.Input("Nombre del Variable Group", "", "", ui.Required)
		})
		if err != nil {
			return err
		}
		description, _ := c.Flags().GetString("description")
		entries, _ := c.Flags().GetStringSlice("variables")

		var vars map[string]models.VariableVal
		if len(entries) == 0 && ui.Interactive() {
			if description == "" {
				if description, err = ui.Input("Descripción (opcional)", "", "", nil); err != nil {
					return err
				}
			}
			if vars, err = cmd.PromptVariables(); err != nil {
				return err
			}
		} else if vars, err = cmd.ParseVariables(entries); err != nil {
			return err
		}

		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		existing, err := vg.GetVariableGroupByName(client, name)
		if err != nil {
			return err
		}
		for _, g := range existing {
			if g.Name == name {
				return fmt.Errorf("el Variable Group '%s' ya existe (ID %d). Usa 'variables update' para modificarlo", name, g.Id)
			}
		}

		created, err := vg.CreateVariableGroup(client, name, vars, description)
		if err != nil {
			return err
		}
		ui.Success("Variable Group '%s' creado (ID: %d) con %d variables", created.Name, created.Id, len(vars))
		ui.Link(client.VariableGroupWebURL(created.Id))
		if cmd.WantsJSON() {
			return ui.PrintJSON(created)
		}
		return nil
	},
}

func init() {
	createVariableGroupCmd.Flags().StringP("name", "n", "", "Nombre del Variable Group")
	createVariableGroupCmd.Flags().StringSliceP("variables", "v", nil, "Variables en formato clave=valor o secret:clave=valor (se puede repetir)")
	createVariableGroupCmd.Flags().StringP("description", "d", "", "Descripción del Variable Group (opcional)")
	cmd.Variables.AddCommand(createVariableGroupCmd)
}
