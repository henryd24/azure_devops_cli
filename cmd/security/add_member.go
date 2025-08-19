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
	Short: "Agrega un usuario a un grupo de seguridad.",
	Run: func(cmd *cobra.Command, args []string) {
		groupName, _ := cmd.Flags().GetString("group")
		user, _ := cmd.Flags().GetString("user")

		if groupName == "" || user == "" {
			log.Fatal("Los flags --group y --user son obligatorios.")
		}

		client := azdevops.GetClientFromEnv()

		group, err := security.SearchGroupByName(client, groupName)
		if err != nil {
			log.Fatalf("Error al buscar el grupo: %v", err)
		}

		if _, err := security.AddUserToGroup(client, group.Descriptor, user); err != nil {
			log.Fatalf("Error al agregar el usuario al grupo: %v", err)
		}

		fmt.Printf("✔ Usuario '%s' agregado con éxito al grupo '%s'.\n", user, group.DisplayName)
	},
}

func init() {
	addMemberCmd.Flags().StringP("group", "g", "", "Nombre del grupo de seguridad (obligatorio)")
	addMemberCmd.Flags().StringP("user", "u", "", "Email o UPN del usuario a agregar (obligatorio)")
	cmd.Security.AddCommand(addMemberCmd)
}
