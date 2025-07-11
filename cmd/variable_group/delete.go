package variable_group

import (
	"fmt"
	"log"

	"azuredevops/azdevops"
	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"

	"github.com/spf13/cobra"
)

var deleteVariableGroupCmd = &cobra.Command{
	Use:   "delete",
	Short: "Elimina un Variable Group por nombre",
	Run: func(cmd *cobra.Command, args []string) {
		groupNames, err := cmd.Flags().GetStringSlice("name")
		if err != nil {
			log.Fatalf("Error al obtener los nombres de los grupos: %v", err)
		}
		if len(groupNames) == 0 {
			log.Fatal("Debes proporcionar al menos un nombre de grupo con --name")
		}

		client := azdevops.GetClientFromEnv()

		success, err := vg.DeleteVariableGroupByName(client, groupNames)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		if !success {
			fmt.Printf("No se encontró el Variable Group con nombre '%s'\n", groupNames)
			return
		}
	},
}

func init() {
	deleteVariableGroupCmd.Flags().StringSliceP("name", "n", nil, "Nombre del variable group o grupos a eliminar")
	cmd.Variables.AddCommand(deleteVariableGroupCmd)
}
