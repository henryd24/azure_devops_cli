package variable_group

import (
	"fmt"
	"strings"

	"azuredevops/azdevops/organization"
	"azuredevops/azdevops/security"
	vg "azuredevops/azdevops/variable_group"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var setPermissionsCmd = &cobra.Command{
	Use:   "set-permissions",
	Short: "Asigna un rol a usuarios/grupos en uno o más Variable Groups",
	Example: `  azdevops variables set-permissions --variable MiGrupo --user ana@empresa.com --group Devs --role Reader
  azdevops variables set-permissions   # modo guiado`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		names, err := cmd.ResolveSliceFlag(c, "variable", func() ([]string, error) {
			return cmd.PickVariableGroupNames(client, "Variable Groups")
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
			u, err := ui.Input("Usuarios (emails separados por coma)", "Opcional", "", nil)
			if err != nil {
				return err
			}
			g, err := ui.Input("Grupos de seguridad del proyecto (separados por coma)", "Sin el prefijo [Proyecto]\\ · Opcional", "", nil)
			if err != nil {
				return err
			}
			users, groups = splitList(u), splitList(g)
			if len(users) == 0 && len(groups) == 0 {
				return fmt.Errorf("debes indicar al menos un usuario o grupo")
			}
		}
		role, err := cmd.ResolveStringFlag(c, "role", func() (string, error) {
			opts := make([]ui.Option[string], len(vg.ValidRoles))
			for i, r := range vg.ValidRoles {
				opts[i] = ui.Option[string]{Label: r, Value: r}
			}
			return ui.Select("Rol", opts)
		})
		if err != nil {
			return err
		}
		if role, err = vg.NormalizeRole(role); err != nil {
			return err
		}

		ui.Info("Buscando identidades...")
		var identityIDs []string
		for _, u := range users {
			id, err := security.GetUserByPrincipalName(client, u)
			if err != nil {
				return fmt.Errorf("operación cancelada: %w", err)
			}
			identityIDs = append(identityIDs, id.ID)
			ui.Success("Usuario: %s", u)
		}
		for _, g := range groups {
			id, err := security.GetGroupByPrincipalName(client, g)
			if err != nil {
				return fmt.Errorf("operación cancelada: %w", err)
			}
			identityIDs = append(identityIDs, id.ID)
			ui.Success("Grupo: %s", g)
		}

		project, err := organization.GetProject(client)
		if err != nil {
			return err
		}
		return forEachGroup(client, names, func(g models.VariableGroup) error {
			if err := vg.SetPermissions(client, project.ID, g.Id, identityIDs, role); err != nil {
				return err
			}
			ui.Success("Rol %s asignado en '%s'", role, g.Name)
			return nil
		})
	},
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func init() {
	setPermissionsCmd.Flags().StringSliceP("variable", "v", nil, "Nombre del Variable Group (se puede repetir)")
	setPermissionsCmd.Flags().StringSliceP("user", "u", nil, "Email del usuario (se puede repetir)")
	setPermissionsCmd.Flags().StringSliceP("group", "g", nil, "Grupo de seguridad del proyecto, sin el prefijo [Proyecto]\\ (se puede repetir)")
	setPermissionsCmd.Flags().StringP("role", "r", "", "Rol a asignar: Reader, User o Administrator")
	_ = setPermissionsCmd.RegisterFlagCompletionFunc("variable", cmd.CompleteVariableGroups)
	_ = setPermissionsCmd.RegisterFlagCompletionFunc("role", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return vg.ValidRoles, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Variables.AddCommand(setPermissionsCmd)
}
