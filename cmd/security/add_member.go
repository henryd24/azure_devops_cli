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
		targetGroups, _ := cmd.Flags().GetStringSlice("target-group")
		groupNames, _ := cmd.Flags().GetStringSlice("group")
		users, _ := cmd.Flags().GetStringSlice("user")

		if len(groupNames) == 0 || len(users) == 0 || len(targetGroups) == 0 {
			log.Fatal("Debes proporcionar al menos un valor para --target-group, --group y --user.")
		}

		client := azdevops.GetClientFromEnv()
		fmt.Println("Buscando descriptores de grupos destino...")
		var targetGroupDescriptors []struct{ Name, Descriptor string }
		for _, targetGroup := range targetGroups {
			group, err := security.GetGroupByPrincipalName(client, targetGroup)
			if err != nil {
				continue
			}
			targetGroupDescriptors = append(targetGroupDescriptors, struct{ Name, Descriptor string }{targetGroup, group.SubjectDescriptor})
		}

		fmt.Println("\nBuscando descriptores de grupos...")
		var groupDescriptors []struct{ Name, Descriptor string }
		for _, groupName := range groupNames {
			group, err := security.GetGroupByPrincipalName(client, groupName)
			if err != nil {
				continue
			}
			groupDescriptors = append(groupDescriptors, struct{ Name, Descriptor string }{groupName, group.SubjectDescriptor})
		}

		fmt.Println("\nBuscando descriptores de usuarios...")
		var userDescriptors []struct{ Name, Descriptor string }
		for _, userPrincipalName := range users {
			user, err := security.GetUserByPrincipalName(client, userPrincipalName)
			if err != nil {
				continue
			}
			userDescriptors = append(userDescriptors, struct{ Name, Descriptor string }{userPrincipalName, user.SubjectDescriptor})
		}
		fmt.Println("\nAsignando membresías...")
		allGroupDescriptors := append(groupDescriptors, userDescriptors...)

		for _, group := range targetGroupDescriptors {
			fmt.Printf("Procesando grupo: '%s'\n", group.Name)
			for _, user := range allGroupDescriptors {
				if _, err := security.AddUserToGroupOptimized(client, group.Descriptor, user.Descriptor); err != nil {
					log.Printf("  └─ ❌ Error al agregar a '%s': %v\n", user.Name, err)
				} else {
					fmt.Printf("  └─ ✔️ Usuario/Grupo '%s' agregado con éxito.\n", user.Name)
				}
			}
		}
	},
}

func init() {
	addMemberCmd.Flags().StringSliceP("target-group", "t", nil, "Nombre del grupo al que se agregara (se puede repetir para múltiples grupos)")
	addMemberCmd.Flags().StringSliceP("user", "u", nil, "Email o UPN del usuario (se puede repetir para múltiples usuarios)")
	addMemberCmd.Flags().StringSliceP("group", "g", nil, "Nombre del grupo a agregar (se puede repetir para múltiples grupos)")
	cmd.Security.AddCommand(addMemberCmd)
}
