package variable_group

import (
	"fmt"
	"log"

	"azuredevops/azdevops"
	"azuredevops/azdevops/security"
	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"

	"github.com/spf13/cobra"
)

var setPermissionsCmd = &cobra.Command{
	Use:   "set-permissions",
	Short: "Asigna permisos a uno o más Variable Groups.",
	Run: func(cmd *cobra.Command, args []string) {
		variableNames, _ := cmd.Flags().GetStringSlice("variable")
		users, _ := cmd.Flags().GetStringSlice("user")
		groups, _ := cmd.Flags().GetStringSlice("group")
		role, _ := cmd.Flags().GetString("role")

		if len(variableNames) == 0 || (len(users) == 0 && len(groups) == 0) || role == "" {
			log.Fatal("Los flags --variable y --role son obligatorios. Debes proporcionar al menos un --user o --group.")
		}

		client := azdevops.GetClientFromEnv()

		var identityIDs []string

		fmt.Println("Buscando identidades de usuarios y grupos...")
		for _, principalName := range users {
			identity, err := security.GetUserByPrincipalName(client, principalName)
			if err != nil {
				log.Fatalf("  └─ ❌ Error fatal: No se pudo encontrar a '%s'. La operación se ha cancelado. Detalles: %v", principalName, err)
			}
			identityIDs = append(identityIDs, identity.ID)
			fmt.Printf("  └─ ✔️ Encontrado: %s\n", principalName)
		}

		for _, group := range groups {
			identity, err := security.GetGroupByPrincipalName(client, group)
			if err != nil {
				log.Fatalf("  └─ ❌ Error fatal: No se pudo encontrar al grupo '%s'. La operación se ha cancelado. Detalles: %v", group, err)
			}
			identityIDs = append(identityIDs, identity.ID)
			fmt.Printf("  └─ ✔️ Encontrado: %s\n", group)
		}

		fmt.Println("\nAplicando permisos a los Variable Groups...")
		for _, variableName := range variableNames {
			vgs, err := vg.GetVariableGroupByName(client, variableName)
			if err != nil || len(vgs) == 0 {
				log.Printf("  └─ ❌ Error al buscar el Variable Group '%s'. Saltando...\n", variableName)
				continue
			}

			variableGroup := vgs[0]

			err = vg.SetPermissionsOptimized(client, &variableGroup, identityIDs, role)
			if err != nil {
				log.Printf("  └─ ❌ Error al asignar permisos en '%s': %v\n", variableName, err)
			} else {
				fmt.Printf("  └─ ✔️ Permisos asignados correctamente en '%s'.\n", variableName)
			}
		}
	},
}

func init() {
	setPermissionsCmd.Flags().StringSliceP("variable", "v", nil, "Nombre del Variable Group (se puede repetir)")
	setPermissionsCmd.Flags().StringSliceP("user", "u", nil, "Email del usuario al que se le asignará el permiso (se puede repetir)")
	setPermissionsCmd.Flags().StringSliceP("group", "g", nil, "Nombre del grupo al que se le asignará el permiso (se puede repetir)")
	setPermissionsCmd.Flags().StringP("role", "r", "", "Rol a asignar: Reader, User, o Administrator")
	cmd.Variables.AddCommand(setPermissionsCmd)
}
