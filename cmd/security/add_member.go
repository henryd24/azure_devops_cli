package security

import (
	"fmt"
	"log"

	"azuredevops/azdevops"
	"azuredevops/azdevops/security"
	"azuredevops/cmd"

	"github.com/spf13/cobra"
)

var addMemberCmd = &cobra.Command{
	Use:   "add-member",
	Short: "Agrega uno o más usuarios a uno o más grupos de seguridad.",
	Run: func(cmd *cobra.Command, args []string) {
		groupNames, _ := cmd.Flags().GetStringSlice("group")
		users, _ := cmd.Flags().GetStringSlice("user")

		if len(groupNames) == 0 || len(users) == 0 {
			log.Fatal("Debes proporcionar al menos un valor para --group y --user.")
		}

		client := azdevops.GetClientFromEnv()

		for _, groupName := range groupNames {
			fmt.Printf("Procesando grupo: '%s'\n", groupName)

			group, err := security.GetGroupByPrincipalName(client, groupName)
			if err != nil {
				log.Printf("  └─ ❌ Error al buscar el grupo '%s': %v\n", groupName, err)
				continue
			}

			for _, user := range users {
				if _, err := security.AddUserToGroup(client, group.SubjectDescriptor, user); err != nil {
					log.Printf("  └─ ❌ Error al agregar a '%s': %v\n", user, err)
				} else {
					fmt.Printf("  └─ ✔️ Usuario '%s' agregado con éxito.\n", user)
				}
			}
		}
	},
}

func init() {
	addMemberCmd.Flags().StringSliceP("group", "g", nil, "Nombre del grupo (se puede repetir para múltiples grupos)")
	addMemberCmd.Flags().StringSliceP("user", "u", nil, "Email o UPN del usuario (se puede repetir para múltiples usuarios)")
	cmd.Security.AddCommand(addMemberCmd)
}
