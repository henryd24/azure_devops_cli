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

var listGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "Lista todos los grupos de seguridad del proyecto.",
	Run: func(cmd *cobra.Command, args []string) {
		client := azdevops.GetClientFromEnv()
		groups, err := security.ListGroups(client)
		if err != nil {
			log.Fatalf("Error al listar los grupos: %v", err)
		}

		jsonData, err := json.MarshalIndent(groups, "", "  ")
		if err != nil {
			log.Fatalf("Error al convertir a JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	},
}

func init() {
	cmd.Security.AddCommand(listGroupsCmd)
}
