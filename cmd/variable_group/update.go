package variable_groups

import (
	"log"
	"strings"

	"azuredevops/azdevops"
	"azuredevops/cmd"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var updateVariableGroupCmd = &cobra.Command{
	Use:   "update",
	Short: "Actualiza un Variable Group por nombre",
	Run: func(cmd *cobra.Command, args []string) {
		groupName, err := cmd.Flags().GetString("name")
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
		groups, err := client.GetVariableGroupByName(groupName)
		if err != nil {
			log.Fatalf("Error al obtener el Variable Group: %v", err)
		}
		if len(groups) > 1 {
			log.Fatalf("Se esperó un solo Variable Group con nombre '%s', pero se encontraron %d", groupName, len(groups))
		}
		if len(groups) == 0 {
			log.Fatalf("No existe un Variable Group con nombre '%s'. Usa el comando 'create' para crearlo.", groupName)
		}

		group := groups[0]
		log.Printf("Actualizando Variable Group: %s (ID: %d)", group.Name, group.Id)
		if _, err := client.AddVariablesToGroup(group, newVariablesMap); err != nil {
			log.Fatalf("Error al agregar variables al grupo: %v", err)
		}
	},
}

func init() {
	updateVariableGroupCmd.Flags().StringP("name", "n", "", "Nombre del variable group")
	updateVariableGroupCmd.Flags().StringSliceP("variables", "v", nil, "Variables a agregar en formato: clave=valor o secret:clave=valor")
	cmd.Variables.AddCommand(updateVariableGroupCmd)
}
