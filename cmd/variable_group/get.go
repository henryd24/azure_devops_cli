package variable_group

import (
	"fmt"

	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var getVariableGroupCmd = &cobra.Command{
	Use:   "get",
	Short: "Obtiene uno o más Variable Groups por nombre o ID",
	Example: `  azdevops variables get --name "MiGrupo"
  azdevops variables get --name "MiGrupo*"
  azdevops variables get --id 42 -o table`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("json")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}

		var groups []models.VariableGroup
		if id, _ := c.Flags().GetInt("id"); id != 0 {
			g, err := vg.GetVariableGroupById(client, id)
			if err != nil {
				return err
			}
			groups = []models.VariableGroup{*g}
		} else {
			name, err := cmd.ResolveStringFlag(c, "name", func() (string, error) {
				return cmd.PickVariableGroupName(client, "Variable Group")
			})
			if err != nil {
				return err
			}
			if groups, err = vg.GetVariableGroupByName(client, name); err != nil {
				return err
			}
			if len(groups) == 0 {
				return fmt.Errorf("no se encontró el Variable Group '%s'", name)
			}
		}

		if out != "table" {
			return cmd.Print(groups)
		}
		for _, g := range groups {
			printGroupVariables(g)
		}
		return nil
	},
}

func init() {
	getVariableGroupCmd.Flags().StringP("name", "n", "", "Nombre del Variable Group (admite comodines)")
	getVariableGroupCmd.Flags().IntP("id", "i", 0, "ID del Variable Group")
	_ = getVariableGroupCmd.RegisterFlagCompletionFunc("name", cmd.CompleteVariableGroups)
	cmd.Variables.AddCommand(getVariableGroupCmd)
}
