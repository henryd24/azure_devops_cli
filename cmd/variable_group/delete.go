package variable_group

import (
	"fmt"
	"sort"
	"strings"

	"azuredevops/azdevops/organization"
	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var deleteVariableGroupCmd = &cobra.Command{
	Use:   "delete",
	Short: "Elimina Variable Groups completos o variables específicas",
	Example: `  azdevops variables delete --name MiGrupo --variables clave1,clave2
  azdevops variables delete --name MiGrupo --yes
  azdevops variables delete   # modo guiado`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		names, err := cmd.ResolveSliceFlag(c, "name", func() ([]string, error) {
			return cmd.PickVariableGroupNames(client, "Variable Groups")
		})
		if err != nil {
			return err
		}
		yes, _ := c.Flags().GetBool("yes")
		variables, _ := c.Flags().GetStringSlice("variables")

		var groups []models.VariableGroup
		for _, name := range names {
			found, err := vg.GetVariableGroupByName(client, name)
			if err != nil {
				return err
			}
			if len(found) == 0 {
				ui.Warn("No se encontró el Variable Group '%s'", name)
			}
			groups = append(groups, found...)
		}
		if len(groups) == 0 {
			return fmt.Errorf("no hay Variable Groups que eliminar")
		}

		if len(variables) == 0 && ui.Interactive() && !yes {
			mode, err := ui.Select("¿Qué quieres eliminar?", []ui.Option[string]{
				{Label: "Algunas variables del grupo", Value: "vars"},
				{Label: "El Variable Group completo", Value: "group"},
			})
			if err != nil {
				return err
			}
			if mode == "vars" {
				if variables, err = pickVariables(groups); err != nil {
					return err
				}
			}
		}

		if len(variables) > 0 {
			if ui.Interactive() {
				if err := cmd.Confirm(fmt.Sprintf("¿Eliminar %s de %d grupo(s)?", strings.Join(variables, ", "), len(groups)), yes); err != nil {
					return err
				}
			}
			failed := 0
			for _, g := range groups {
				missing, err := vg.RemoveVariablesFromGroup(client, g, variables)
				if err != nil {
					ui.Errorf("%s: %v", g.Name, err)
					failed++
					continue
				}
				for _, m := range missing {
					ui.Warn("La variable '%s' no existe en '%s'", m, g.Name)
				}
				if len(missing) < len(variables) {
					ui.Success("Variables eliminadas de '%s' (ID %d)", g.Name, g.Id)
					ui.Link(client.VariableGroupWebURL(g.Id))
				}
			}
			if failed > 0 {
				return fmt.Errorf("%d grupo(s) no se pudieron actualizar", failed)
			}
			return nil
		}

		list := make([]string, len(groups))
		for i, g := range groups {
			list[i] = fmt.Sprintf("%s (ID %d)", g.Name, g.Id)
		}
		ui.Warn("Se eliminarán: %s", strings.Join(list, ", "))
		if err := cmd.Confirm(fmt.Sprintf("¿Eliminar %d Variable Group(s) de forma permanente?", len(groups)), yes); err != nil {
			return err
		}

		project, err := organization.GetProject(client)
		if err != nil {
			return err
		}
		failed := 0
		for _, g := range groups {
			if err := vg.DeleteVariableGroup(client, g.Id, project.ID); err != nil {
				ui.Errorf("%v", err)
				failed++
				continue
			}
			ui.Success("Variable Group '%s' eliminado (ID: %d)", g.Name, g.Id)
		}
		if failed > 0 {
			return fmt.Errorf("%d Variable Group(s) no se pudieron eliminar", failed)
		}
		return nil
	},
}

func pickVariables(groups []models.VariableGroup) ([]string, error) {
	seen := map[string]bool{}
	for _, g := range groups {
		for k := range g.Variables {
			seen[k] = true
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	opts := make([]ui.Option[string], len(keys))
	for i, k := range keys {
		opts[i] = ui.Option[string]{Label: k, Value: k}
	}
	return ui.MultiSelect("Variables a eliminar", opts)
}

func init() {
	deleteVariableGroupCmd.Flags().StringSliceP("name", "n", nil, "Nombre del Variable Group (se puede repetir, admite comodines)")
	deleteVariableGroupCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
	deleteVariableGroupCmd.Flags().StringSliceP("variables", "v", nil, "Variables específicas a eliminar (si se omite se elimina el grupo completo)")
	_ = deleteVariableGroupCmd.RegisterFlagCompletionFunc("name", cmd.CompleteVariableGroups)
	cmd.Variables.AddCommand(deleteVariableGroupCmd)
}
