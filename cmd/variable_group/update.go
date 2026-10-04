package variable_group

import (
	"fmt"

	"azuredevops/azdevops"
	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var updateVariableGroupCmd = &cobra.Command{
	Use:   "update",
	Short: "Agrega o modifica variables en uno o más Variable Groups",
	Example: `  azdevops variables update --name MiGrupo -v clave2=nuevo -v secret:token=xyz
  azdevops variables update --name MiGrupo --name OtroGrupo -v "a=1,b=2"
  azdevops variables update   # modo guiado`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		names, err := cmd.ResolveSliceFlag(c, "name", func() ([]string, error) {
			return cmd.PickVariableGroupNames(client, "Variable Groups a actualizar")
		})
		if err != nil {
			return err
		}
		description, _ := c.Flags().GetString("description")
		entries, _ := c.Flags().GetStringSlice("variables")

		var vars map[string]models.VariableVal
		if len(entries) == 0 && description == "" && ui.Interactive() {
			if vars, err = cmd.PromptVariables(); err != nil {
				return err
			}
		} else if vars, err = cmd.ParseVariables(entries); err != nil {
			return err
		}
		if len(vars) == 0 && description == "" {
			return fmt.Errorf("no hay nada que actualizar: indica --variables o --description")
		}

		return forEachGroup(client, names, func(g models.VariableGroup) error {
			updated, err := vg.AddVariablesToGroup(client, g, vars, description)
			if err != nil {
				return err
			}
			ui.Success("Variable Group '%s' (ID %d) actualizado: %d variables agregadas/modificadas", updated.Name, updated.Id, len(vars))
			ui.Link(client.VariableGroupWebURL(updated.Id))
			return nil
		})
	},
}

// forEachGroup resuelve cada nombre (admite comodines) y aplica fn a cada grupo encontrado.
// Continúa con los demás grupos si uno falla y devuelve un error resumen al final.
func forEachGroup(client *azdevops.Client, names []string, fn func(models.VariableGroup) error) error {
	failed := 0
	for _, name := range names {
		groups, err := vg.GetVariableGroupByName(client, name)
		if err != nil {
			return err
		}
		if len(groups) == 0 {
			ui.Warn("No se encontró el Variable Group '%s'", name)
			failed++
			continue
		}
		for _, g := range groups {
			if err := fn(g); err != nil {
				ui.Errorf("%s: %v", g.Name, err)
				failed++
			}
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d operación(es) fallaron", failed)
	}
	return nil
}

func init() {
	updateVariableGroupCmd.Flags().StringSliceP("name", "n", nil, "Nombre del Variable Group (se puede repetir, admite comodines)")
	updateVariableGroupCmd.Flags().StringSliceP("variables", "v", nil, "Variables en formato clave=valor o secret:clave=valor")
	updateVariableGroupCmd.Flags().StringP("description", "d", "", "Nueva descripción (opcional)")
	_ = updateVariableGroupCmd.RegisterFlagCompletionFunc("name", cmd.CompleteVariableGroups)
	cmd.Variables.AddCommand(updateVariableGroupCmd)
}
