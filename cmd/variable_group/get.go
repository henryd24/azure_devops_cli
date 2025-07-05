package variable_groups

import (
	"encoding/json"
	"fmt"
	"log"

	"azurevgcli/azdevops"
	"azurevgcli/cmd"

	"github.com/spf13/cobra"
)

var getVariableGroupCmd = &cobra.Command{
	Use:   "get",
	Short: "Obtiene un Variable Group por nombre",
	Run: func(cmd *cobra.Command, args []string) {
		groupName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("Error al obtener el nombre del grupo: %v", err)
		}
		if groupName == "" {
			log.Fatal("Debes proporcionar el nombre del grupo con --name")
		}

		client := azdevops.GetClientFromEnv()

		group, err := client.GetVariableGroupByName(groupName)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		if group == nil {
			fmt.Printf("No se encontró el Variable Group con nombre '%s'\n", groupName)
			return
		}

		jsonData, err := json.Marshal(group)
		if err != nil {
			log.Fatalf("Error al convertir a JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	},
}

func init() {
	getVariableGroupCmd.Flags().StringP("name", "n", "", "Nombre del variable group")
	cmd.Variables.AddCommand(getVariableGroupCmd)
}
