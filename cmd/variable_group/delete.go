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
	Short: "Elimina un Variable Group completo o variables específicas de uno",
	Run: func(cmd *cobra.Command, args []string) {
		groupNames, err := cmd.Flags().GetStringSlice("name")
		if err != nil {
			log.Fatalf("Error al obtener los nombres del grupo: %v", err)
		}
		if len(groupNames) == 0 {
			log.Fatal("Debes proporcionar al menos un nombre de grupo con --name")
		}

		variablesToDelete, err := cmd.Flags().GetStringSlice("variables")
		if err != nil {
			log.Fatalf("Error al obtener las variables: %v", err)
		}

		client := azdevops.GetClientFromEnv()

		if len(variablesToDelete) > 0 {
			if _, err := vg.DeleteVariableInGroupById(client, groupNames, variablesToDelete); err != nil {
				log.Fatalf("Error al eliminar variables del grupo: %v", err)
			}
			return
		}

		autoconfirmation, _ := cmd.Flags().GetBool("yes")
		if !autoconfirmation {
			fmt.Printf("Estás a punto de eliminar los siguientes Variable Groups: %v\n", groupNames)
			fmt.Print("¿Estás seguro de que deseas continuar? (s/n): ")
			var confirmation string
			fmt.Scanln(&confirmation)
			if confirmation != "s" && confirmation != "S" {
				fmt.Println("Operación cancelada.")
				return
			}
		}

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
	deleteVariableGroupCmd.Flags().BoolP("yes", "y", false, "Confirmar eliminación sin preguntar")
	deleteVariableGroupCmd.Flags().StringSliceP("variables", "v", nil, "Variables específicas a eliminar del variable group")
	cmd.Variables.AddCommand(deleteVariableGroupCmd)
}
