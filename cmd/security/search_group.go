package security

import (
	"encoding/json"
	"fmt"
	"log"

	"azuredevops/azdevops"
	"azuredevops/azdevops/security"
	"azuredevops/cmd"

	"github.com/spf13/cobra"
)

var searchGroupCmd = &cobra.Command{
	Use:   "search-group",
	Short: "Busca un grupo de seguridad específico por su nombre.",
	Long:  "Busca un grupo de seguridad por su nombre exacto o parcial y devuelve un error si no se encuentra o si hay múltiples coincidencias.",
	Run: func(cmd *cobra.Command, args []string) {
		groupName, _ := cmd.Flags().GetString("name")
		if groupName == "" {
			log.Fatal("El flag --name es obligatorio.")
		}

		client := azdevops.GetClientFromEnv()

		group, err := security.SearchGroupByName(client, groupName)
		if err != nil {
			log.Fatalf("Error al buscar el grupo: %v", err)
		}

		jsonData, err := json.MarshalIndent(group, "", "  ")
		if err != nil {
			log.Fatalf("Error al convertir el resultado a JSON: %v", err)
		}

		fmt.Println(string(jsonData))
	},
}

func init() {
	searchGroupCmd.Flags().StringP("name", "n", "", "Nombre del grupo a buscar (obligatorio)")
	cmd.Security.AddCommand(searchGroupCmd)
}
