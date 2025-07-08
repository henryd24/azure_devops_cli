package variable_groups

import (
	"log"
	"strings"

	"azuredevops/azdevops"
	"azuredevops/cmd"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var createVariableGroupCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new variable group",
	Run: func(cmd *cobra.Command, args []string) {
		groupName, err := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		if err != nil {
			log.Fatalf("Error al obtener el nombre del grupo: %v", err)
		}

		newVariables, err := cmd.Flags().GetStringSlice("variables")
		if err != nil {
			log.Fatalf("Error al obtener las variables: %v", err)
		}

		if groupName == "" {
			log.Fatal("Debes proporcionar el nombre del grupo con --name")
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
		getVariableGroup, err := client.GetVariableGroupByName(groupName)
		if err == nil && len(getVariableGroup) > 0 {
			log.Fatalf("El Variable Group '%s' ya existe. Usa otro nombre.", groupName)
		} else if err != nil && !strings.Contains(err.Error(), "Variable group not found") {
			log.Fatalf("Error al obtener el Variable Group: %v", err)
		}
		_, outputErr := client.CreateVariableGroup(groupName, newVariablesMap, description)
		if outputErr != nil {
			log.Fatalf("Error al agregar variables al grupo: %v", outputErr)
		}
	},
}

func init() {
	createVariableGroupCmd.Flags().StringP("name", "n", "", "Name of the variable group")
	createVariableGroupCmd.Flags().StringSliceP("variables", "v", []string{}, "Variables in the format key=value or secret:key=value")
	createVariableGroupCmd.Flags().StringP("description", "d", "", "Description of the variable group (optional)")
	cmd.Variables.AddCommand(createVariableGroupCmd)
}
