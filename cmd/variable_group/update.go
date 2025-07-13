package variable_group

import (
	"log"
	"strings"

	"azuredevops/azdevops"
	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"
	"azuredevops/models"
	"github.com/spf13/cobra"
)

var updateVariableGroupCmd = &cobra.Command{
	Use:   "update",
	Short: "Actualiza un Variable Group por nombre",
	Run: func(cmd *cobra.Command, args []string) {
		groupNames, err := cmd.Flags().GetStringSlice("name")
		if err != nil {
			log.Fatalf("Error al obtener los nombres del grupo: %v", err)
		}
		if len(groupNames) == 0 {
			log.Fatal("Debes proporcionar al menos un nombre de grupo con --name")
		}

		description, _ := cmd.Flags().GetString("description")

		newVariables, err := cmd.Flags().GetStringSlice("variables")
		if err != nil {
			log.Fatalf("Error al obtener las variables: %v", err)
		}

		client := azdevops.GetClientFromEnv()
		newVariablesMap := make(map[string]models.VariableVal)

		for _, v := range newVariables {
			isSecret := false
			variable := v

			if strings.HasPrefix(v, "secret:") {
				isSecret = true
				variable = strings.TrimPrefix(v, "secret:")
			}

			parts := strings.SplitN(variable, "=", 2)
			if len(parts) != 2 {
				log.Fatalf("Formato de variable inválido: %s. Usa clave=valor o secret:clave=valor", v)
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if key == "" {
				log.Fatalf("La clave no puede estar vacía en: %s", v)
			}

			newVariablesMap[key] = models.VariableVal{
				Value:    value,
				IsSecret: isSecret,
			}
		}

		if _, err := vg.AddVariablesToGroup(client, groupNames, newVariablesMap, description); err != nil {
			log.Fatalf("Error al agregar variables al grupo: %v", err)
		}
	},
}

func init() {
	updateVariableGroupCmd.Flags().StringSliceP("name", "n", nil, "Nombre del variable group")
	updateVariableGroupCmd.Flags().StringSliceP("variables", "v", nil, "Variables a agregar en formato: clave=valor o secret:clave=valor")
	updateVariableGroupCmd.Flags().StringP("description", "d", "", "Descripción del variable group (opcional)")
	cmd.Variables.AddCommand(updateVariableGroupCmd)
}
