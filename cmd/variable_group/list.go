package variable_group

import (
	"fmt"
	"strconv"

	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var listVariableGroupsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lista los Variable Groups del proyecto",
	Example: `  azdevops variables list
  azdevops variables list --filter "app-*" -o json`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		filter, _ := c.Flags().GetString("filter")
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		var groups []models.VariableGroup
		if err := ui.Spinner("Cargando Variable Groups...", func() (err error) {
			groups, err = vg.ListVariableGroups(client, filter)
			return err
		}); err != nil {
			return err
		}
		return printGroups(groups, out)
	},
}

func printGroups(groups []models.VariableGroup, out string) error {
	if out != "table" {
		if groups == nil {
			groups = []models.VariableGroup{}
		}
		return cmd.Print(groups)
	}
	if len(groups) == 0 {
		ui.Info("No se encontraron Variable Groups.")
		return nil
	}
	rows := make([][]string, len(groups))
	for i, g := range groups {
		secrets := 0
		for _, v := range g.Variables {
			if v.IsSecret {
				secrets++
			}
		}
		rows[i] = []string{
			strconv.Itoa(g.Id), g.Name,
			fmt.Sprintf("%d (%d secretas)", len(g.Variables), secrets),
			g.ModifiedBy.DisplayName, ui.FormatTime(&g.ModifiedOn), truncate(g.Description, 40),
		}
	}
	ui.Table([]string{"id", "nombre", "variables", "modificado por", "modificado", "descripción"}, rows)
	return nil
}

func printGroupVariables(g models.VariableGroup) {
	fmt.Fprintf(ui.Out, "%s (ID %d)\n", ui.Bold(g.Name), g.Id)
	if g.Description != "" {
		fmt.Fprintf(ui.Out, "%s\n", g.Description)
	}
	rows := make([][]string, 0, len(g.Variables))
	for _, k := range sortedKeys(g.Variables) {
		v := g.Variables[k]
		value := v.Value
		if v.IsSecret {
			value = "********"
		}
		rows = append(rows, []string{k, value})
	}
	ui.Table([]string{"variable", "valor"}, rows)
	fmt.Fprintln(ui.Out)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func init() {
	listVariableGroupsCmd.Flags().StringP("filter", "f", "", "Filtro por nombre, admite comodines (p. ej. 'app-*')")
	cmd.Variables.AddCommand(listVariableGroupsCmd)
}
