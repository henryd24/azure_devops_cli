package security

import (
	"fmt"

	"azuredevops/azdevops/organization"
	"azuredevops/azdevops/security"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var listGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "Lista los grupos de seguridad (de la organización o del proyecto)",
	Example: `  azdevops security list-groups --search devs -o table
  azdevops security list-groups --project-only`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("json")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		search, _ := c.Flags().GetString("search")
		projectOnly, _ := c.Flags().GetBool("project-only")

		var groups []models.GraphGroup
		if err := ui.Spinner("Cargando grupos...", func() error {
			scope := ""
			if projectOnly {
				if scope, err = organization.GetProjectDescriptor(client); err != nil {
					return err
				}
			}
			groups, err = security.ListGroups(client, scope, search)
			return err
		}); err != nil {
			return err
		}
		if out != "table" {
			if groups == nil {
				groups = []models.GraphGroup{}
			}
			return cmd.Print(groups)
		}
		rows := make([][]string, len(groups))
		for i, g := range groups {
			rows[i] = []string{g.PrincipalName, g.Description}
		}
		ui.Table([]string{"grupo", "descripción"}, rows)
		return nil
	},
}

var searchGroupCmd = &cobra.Command{
	Use:   "search-group",
	Short: "Busca un grupo de seguridad del proyecto por su nombre",
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		name, err := cmd.ResolveStringFlag(c, "name", func() (string, error) {
			names, err := pickProjectGroups(client, "Grupo", false)
			if err != nil {
				return "", err
			}
			return names[0], nil
		})
		if err != nil {
			return err
		}
		group, err := security.GetGroupByPrincipalName(client, name)
		if err != nil {
			return err
		}
		return cmd.Print(group)
	},
}

var listMembersCmd = &cobra.Command{
	Use:   "list-members",
	Short: "Lista los miembros directos de un grupo del proyecto",
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		name, err := cmd.ResolveStringFlag(c, "group", func() (string, error) {
			names, err := pickProjectGroups(client, "Grupo", false)
			if err != nil {
				return "", err
			}
			return names[0], nil
		})
		if err != nil {
			return err
		}
		group, err := security.GetGroupByPrincipalName(client, name)
		if err != nil {
			return err
		}
		members, err := security.ListMembers(client, group.SubjectDescriptor)
		if err != nil {
			return err
		}
		if out != "table" {
			if members == nil {
				members = []models.GraphSubject{}
			}
			return cmd.Print(members)
		}
		if len(members) == 0 {
			ui.Info("El grupo '%s' no tiene miembros directos", name)
			return nil
		}
		rows := make([][]string, len(members))
		for i, m := range members {
			id := m.MailAddress
			if id == "" {
				id = m.PrincipalName
			}
			rows[i] = []string{m.DisplayName, id, m.SubjectKind}
		}
		ui.Table([]string{"nombre", "identificador", "tipo"}, rows)
		fmt.Fprintf(ui.Err, "%d miembro(s)\n", len(members))
		return nil
	},
}

func init() {
	listGroupsCmd.Flags().StringP("search", "s", "", "Filtra por nombre (sin distinguir mayúsculas)")
	listGroupsCmd.Flags().Bool("project-only", false, "Solo grupos del proyecto actual")
	searchGroupCmd.Flags().StringP("name", "n", "", "Nombre del grupo, sin el prefijo [Proyecto]\\")
	listMembersCmd.Flags().StringP("group", "g", "", "Nombre del grupo, sin el prefijo [Proyecto]\\")
	cmd.Security.AddCommand(listGroupsCmd, searchGroupCmd, listMembersCmd)
}
