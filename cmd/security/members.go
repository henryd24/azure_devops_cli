package security

import (
	"fmt"

	"azuredevops/azdevops/security"
	"azuredevops/cmd"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var addMemberCmd = &cobra.Command{
	Use:   "add-member",
	Short: "Agrega usuarios y/o grupos a uno o más grupos de seguridad",
	Example: `  azdevops security add-member --target-group Destino --user ana@empresa.com --group Devs
  azdevops security add-member   # modo guiado`,
	RunE: func(c *cobra.Command, args []string) error {
		return changeMembership(c, true)
	},
}

var removeMemberCmd = &cobra.Command{
	Use:     "remove-member",
	Short:   "Quita usuarios y/o grupos de uno o más grupos de seguridad",
	Example: `  azdevops security remove-member --target-group Destino --user ana@empresa.com`,
	RunE: func(c *cobra.Command, args []string) error {
		return changeMembership(c, false)
	},
}

func changeMembership(c *cobra.Command, add bool) error {
	client, err := cmd.NewClient()
	if err != nil {
		return err
	}
	targets, err := cmd.ResolveSliceFlag(c, "target-group", func() ([]string, error) {
		return pickProjectGroups(client, "Grupos destino", true)
	})
	if err != nil {
		return err
	}
	users, _ := c.Flags().GetStringSlice("user")
	groups, _ := c.Flags().GetStringSlice("group")
	if len(users) == 0 && len(groups) == 0 {
		if !ui.Interactive() {
			return fmt.Errorf("debes proporcionar al menos un --user o --group")
		}
		if users, groups, err = promptMembers(client); err != nil {
			return err
		}
	}

	targetSubjects := resolveGroups(client, targets)
	members := append(resolveGroups(client, groups), resolveUsers(client, users)...)
	if len(targetSubjects) == 0 || len(members) == 0 {
		return fmt.Errorf("no se encontraron grupos destino o miembros válidos")
	}

	if !add {
		yes, _ := c.Flags().GetBool("yes")
		if err := cmd.Confirm(fmt.Sprintf("¿Quitar %d miembro(s) de %d grupo(s)?", len(members), len(targetSubjects)), yes); err != nil {
			return err
		}
	}

	failed := 0
	for _, target := range targetSubjects {
		ui.Info("Grupo '%s'", target.Name)
		for _, m := range members {
			var err error
			if add {
				err = security.AddMembership(client, target.Descriptor, m.Descriptor)
			} else {
				err = security.RemoveMembership(client, target.Descriptor, m.Descriptor)
			}
			if err != nil {
				ui.Errorf("  %s: %v", m.Name, err)
				failed++
				continue
			}
			if add {
				ui.Success("  '%s' agregado", m.Name)
			} else {
				ui.Success("  '%s' quitado", m.Name)
			}
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d operación(es) fallaron", failed)
	}
	return nil
}

func init() {
	for _, c := range []*cobra.Command{addMemberCmd, removeMemberCmd} {
		c.Flags().StringSliceP("target-group", "t", nil, "Grupo destino, sin el prefijo [Proyecto]\\ (se puede repetir)")
		c.Flags().StringSliceP("user", "u", nil, "Email o UPN del usuario (se puede repetir)")
		c.Flags().StringSliceP("group", "g", nil, "Grupo miembro (se puede repetir)")
	}
	removeMemberCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
	cmd.Security.AddCommand(addMemberCmd, removeMemberCmd)
}
