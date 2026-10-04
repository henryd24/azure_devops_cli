// Package securefiles contiene los comandos de Library > Secure files.
package securefiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"azuredevops/azdevops"
	"azuredevops/azdevops/organization"
	"azuredevops/azdevops/pipeline"
	"azuredevops/azdevops/securefile"
	"azuredevops/azdevops/security"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var secureFilesCmd = &cobra.Command{
	Use:     "securefiles",
	Aliases: []string{"sf", "secure-files"},
	Short:   "Gestiona archivos seguros (certificados, keystores, kubeconfig…)",
	Long: `Gestiona los archivos seguros de Library > Secure files.

Los archivos seguros se guardan cifrados y solo los pipelines autorizados pueden
descargarlos (tarea DownloadSecureFile). La API no permite descargarlos ni cambiar
su contenido; para actualizar uno usa 'securefiles replace', que conserva sus
pipelines autorizados, roles y aprobaciones.`,
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lista los archivos seguros del proyecto",
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		files, err := securefile.List(client)
		if err != nil {
			return err
		}
		if out != "table" {
			if files == nil {
				files = []models.SecureFile{}
			}
			return cmd.Print(files)
		}
		if len(files) == 0 {
			ui.Info("El proyecto no tiene archivos seguros.")
			return nil
		}
		rows := make([][]string, len(files))
		for i, f := range files {
			rows[i] = []string{f.Name, identity(f.ModifiedBy), ui.FormatTime(f.ModifiedOn), f.ID}
		}
		ui.Table([]string{"nombre", "modificado por", "modificado", "id"}, rows)
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Muestra un archivo seguro: pipelines autorizados, roles y aprobaciones",
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		file, err := resolveFile(c, client, "Archivo seguro")
		if err != nil {
			return err
		}
		project, err := organization.GetProject(client)
		if err != nil {
			return err
		}
		snap, err := securefile.TakeSnapshot(client, project.ID, *file)
		if err != nil {
			return err
		}
		if out != "table" {
			return cmd.Print(snap)
		}
		printSnapshot(client, snap)
		return nil
	},
}

func printSnapshot(client *azdevops.Client, snap *securefile.Snapshot) {
	f := snap.File
	fmt.Fprintf(ui.Out, "%s  (%s)\n", ui.Bold(f.Name), f.ID)
	fmt.Fprintf(ui.Out, "  Creado:      %s por %s\n", ui.FormatTime(f.CreatedOn), identity(f.CreatedBy))
	fmt.Fprintf(ui.Out, "  Modificado:  %s por %s\n", ui.FormatTime(f.ModifiedOn), identity(f.ModifiedBy))
	for k, v := range f.Properties {
		fmt.Fprintf(ui.Out, "  Propiedad:   %s=%s\n", k, v)
	}

	fmt.Fprintf(ui.Out, "\n%s\n", ui.Bold("Pipelines autorizados"))
	if snap.Permissions.OpenToAllPipelines() {
		fmt.Fprintln(ui.Out, "  Todos los pipelines (acceso abierto)")
	}
	ids := snap.Permissions.AuthorizedPipelineIDs()
	names := pipelineNames(client)
	for _, id := range ids {
		fmt.Fprintf(ui.Out, "  #%d %s\n", id, names[id])
	}
	if len(ids) == 0 && !snap.Permissions.OpenToAllPipelines() {
		fmt.Fprintln(ui.Out, "  (ninguno)")
	}

	fmt.Fprintf(ui.Out, "\n%s", ui.Bold("Roles"))
	if !snap.Inherits() {
		fmt.Fprint(ui.Out, "  (herencia del proyecto desactivada)")
	}
	fmt.Fprintln(ui.Out)
	rows := make([][]string, len(snap.Roles))
	for i, r := range snap.Roles {
		access := "asignado"
		if r.Access == "inherited" {
			access = "heredado"
		}
		rows[i] = []string{identity(&r.Identity), r.Role.Name, access}
	}
	ui.Table([]string{"identidad", "rol", "origen"}, rows)

	if len(snap.Checks) > 0 {
		fmt.Fprintf(ui.Out, "\n%s\n", ui.Bold("Aprobaciones y checks"))
		for _, ch := range snap.Checks {
			fmt.Fprintf(ui.Out, "  %s\n", ch.Type.Name)
		}
	}
	ui.Link(client.WebURL("_library?itemType=SecureFiles"))
}

var uploadCmd = &cobra.Command{
	Use:   "upload <archivo>",
	Short: "Sube un archivo seguro nuevo",
	Example: `  azdevops securefiles upload ./cert.pfx
  azdevops sf upload ./kubeconfig --name kubeconfig-prod --pipeline 12 --pipeline 15
  azdevops sf upload ./npmrc --authorize-all-pipelines`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		path, err := localPath(args, "Ruta del archivo a subir")
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name, _ := c.Flags().GetString("name")
		if name == "" {
			name = filepath.Base(path)
		}
		if existing, _ := securefile.Find(client, name); existing != nil {
			return fmt.Errorf("ya existe un archivo seguro llamado '%s'; usa 'securefiles replace --name %q %s' para reemplazarlo", name, name, path)
		}
		all, _ := c.Flags().GetBool("authorize-all-pipelines")
		pipelines, _ := c.Flags().GetIntSlice("pipeline")

		var created *models.SecureFile
		if err := ui.Spinner("Subiendo...", func() (err error) {
			created, err = securefile.Upload(client, name, data, all)
			return err
		}); err != nil {
			return err
		}
		ui.Success("Archivo seguro '%s' subido (%s, %d bytes)", created.Name, created.ID, len(data))
		if len(pipelines) > 0 {
			if err := securefile.SetPipelinePermissions(client, created.ID, pipelines, true, nil); err != nil {
				return err
			}
			ui.Success("Autorizado para %d pipeline(s)", len(pipelines))
		}
		if cmd.WantsData() {
			return cmd.Print(created)
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Elimina un archivo seguro",
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		file, err := resolveFile(c, client, "Archivo seguro a eliminar")
		if err != nil {
			return err
		}
		yes, _ := c.Flags().GetBool("yes")
		if err := cmd.Confirm(fmt.Sprintf("¿Eliminar el archivo seguro '%s'? Los pipelines que lo usan fallarán", file.Name), yes); err != nil {
			return err
		}
		if err := securefile.Delete(client, file.ID); err != nil {
			return err
		}
		ui.Success("Archivo seguro '%s' eliminado", file.Name)
		return nil
	},
}

var authorizeCmd = &cobra.Command{
	Use:   "authorize",
	Short: "Autoriza o revoca el uso de un archivo seguro por pipelines",
	Example: `  azdevops sf authorize --name cert.pfx --pipeline 12 --pipeline 15
  azdevops sf authorize --name cert.pfx --pipeline 12 --revoke
  azdevops sf authorize --name npmrc --all-pipelines=true`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		file, err := resolveFile(c, client, "Archivo seguro")
		if err != nil {
			return err
		}
		ids, _ := c.Flags().GetIntSlice("pipeline")
		revoke, _ := c.Flags().GetBool("revoke")
		var all *bool
		if c.Flags().Changed("all-pipelines") {
			v, _ := c.Flags().GetBool("all-pipelines")
			all = &v
		}
		if len(ids) == 0 && all == nil {
			if !ui.Interactive() {
				return fmt.Errorf("indica --pipeline o --all-pipelines")
			}
			id, err := cmd.PickPipeline(client, "Pipeline a autorizar")
			if err != nil {
				return err
			}
			ids = []int{id}
		}
		if err := securefile.SetPipelinePermissions(client, file.ID, ids, !revoke, all); err != nil {
			return err
		}
		verb := "autorizado(s)"
		if revoke {
			verb = "revocado(s)"
		}
		if len(ids) > 0 {
			ui.Success("%d pipeline(s) %s en '%s'", len(ids), verb, file.Name)
		}
		if all != nil {
			ui.Success("Acceso para todos los pipelines: %v", *all)
		}
		return nil
	},
}

var setRoleCmd = &cobra.Command{
	Use:   "set-role",
	Short: "Asigna un rol (Reader, User, Administrator) a usuarios o grupos",
	Example: `  azdevops sf set-role --name cert.pfx --user ana@empresa.com --role Administrator
  azdevops sf set-role --name cert.pfx --group Devs --role User`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		file, err := resolveFile(c, client, "Archivo seguro")
		if err != nil {
			return err
		}
		users, _ := c.Flags().GetStringSlice("user")
		groups, _ := c.Flags().GetStringSlice("group")
		if len(users) == 0 && len(groups) == 0 {
			if !ui.Interactive() {
				return fmt.Errorf("indica al menos un --user o --group")
			}
			u, err := ui.Input("Usuarios (emails separados por coma)", "Opcional", "", nil)
			if err != nil {
				return err
			}
			g, err := ui.Input("Grupos del proyecto (separados por coma)", "Sin el prefijo [Proyecto]\\ · Opcional", "", nil)
			if err != nil {
				return err
			}
			users, groups = splitList(u), splitList(g)
		}
		role, err := cmd.ResolveStringFlag(c, "role", func() (string, error) {
			opts := make([]ui.Option[string], len(securefile.ValidRoles))
			for i, r := range securefile.ValidRoles {
				opts[i] = ui.Option[string]{Label: r, Value: r}
			}
			return ui.Select("Rol", opts)
		})
		if err != nil {
			return err
		}
		role, err = normalizeRole(role)
		if err != nil {
			return err
		}

		var assignments []models.SecurityRoleAssignment
		for _, u := range users {
			id, err := security.GetUserByPrincipalName(client, u)
			if err != nil {
				return err
			}
			assignments = append(assignments, models.SecurityRoleAssignment{RoleName: role, UserID: id.ID})
		}
		for _, g := range groups {
			id, err := security.GetGroupByPrincipalName(client, g)
			if err != nil {
				return err
			}
			assignments = append(assignments, models.SecurityRoleAssignment{RoleName: role, UserID: id.ID})
		}
		if len(assignments) == 0 {
			return fmt.Errorf("no se indicaron usuarios ni grupos")
		}
		project, err := organization.GetProject(client)
		if err != nil {
			return err
		}
		if err := securefile.SetRoles(client, project.ID, file.ID, assignments); err != nil {
			return err
		}
		ui.Success("Rol %s asignado a %d identidad(es) en '%s'", role, len(assignments), file.Name)
		return nil
	},
}

func resolveFile(c *cobra.Command, client *azdevops.Client, title string) (*models.SecureFile, error) {
	if name, _ := c.Flags().GetString("name"); name != "" {
		return securefile.Find(client, name)
	}
	if !ui.Interactive() {
		return nil, ui.MissingError("name")
	}
	files, err := securefile.List(client)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("el proyecto no tiene archivos seguros")
	}
	opts := make([]ui.Option[int], len(files))
	for i, f := range files {
		opts[i] = ui.Option[int]{Label: fmt.Sprintf("%s  · %s", f.Name, ui.FormatTime(f.ModifiedOn)), Value: i}
	}
	idx, err := ui.Select(title, opts)
	if err != nil {
		return nil, err
	}
	return &files[idx], nil
}

func localPath(args []string, title string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if !ui.Interactive() {
		return "", fmt.Errorf("indica la ruta del archivo local")
	}
	return ui.Input(title, "", "", func(s string) error {
		if st, err := os.Stat(strings.TrimSpace(s)); err != nil || st.IsDir() {
			return fmt.Errorf("no existe el archivo")
		}
		return nil
	})
}

func pipelineNames(client *azdevops.Client) map[int]string {
	names := map[int]string{}
	defs, err := pipeline.ListDefinitions(client, "", "")
	if err != nil {
		return names
	}
	for _, d := range defs {
		if d.ID != nil {
			names[int(*d.ID)] = cmd.Str(d.Name)
		}
	}
	return names
}

func identity(id *models.IdentityRef) string {
	if id == nil {
		return "-"
	}
	if id.DisplayName != "" {
		return id.DisplayName
	}
	return id.UniqueName
}

func normalizeRole(role string) (string, error) {
	for _, r := range securefile.ValidRoles {
		if strings.EqualFold(r, role) {
			return r, nil
		}
	}
	return "", fmt.Errorf("rol '%s' no válido (usa %s)", role, strings.Join(securefile.ValidRoles, ", "))
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
	for _, c := range []*cobra.Command{getCmd, deleteCmd, authorizeCmd, setRoleCmd} {
		c.Flags().StringP("name", "n", "", "Nombre (o ID) del archivo seguro")
	}
	uploadCmd.Flags().StringP("name", "n", "", "Nombre en Azure DevOps (por defecto el del archivo)")
	uploadCmd.Flags().Bool("authorize-all-pipelines", false, "Permitir que todos los pipelines lo usen")
	uploadCmd.Flags().IntSlice("pipeline", nil, "ID de pipeline a autorizar (se puede repetir)")
	deleteCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
	authorizeCmd.Flags().IntSlice("pipeline", nil, "ID de pipeline (se puede repetir)")
	authorizeCmd.Flags().Bool("revoke", false, "Revocar en lugar de autorizar")
	authorizeCmd.Flags().Bool("all-pipelines", false, "Abrir (true) o cerrar (false) el acceso a todos los pipelines")
	setRoleCmd.Flags().StringSliceP("user", "u", nil, "Email del usuario (se puede repetir)")
	setRoleCmd.Flags().StringSliceP("group", "g", nil, "Grupo del proyecto sin el prefijo [Proyecto]\\ (se puede repetir)")
	setRoleCmd.Flags().StringP("role", "r", "", "Rol: Reader, User o Administrator")

	secureFilesCmd.AddCommand(listCmd, getCmd, uploadCmd, deleteCmd, authorizeCmd, setRoleCmd)
	cmd.RootCmd.AddCommand(secureFilesCmd)
}
