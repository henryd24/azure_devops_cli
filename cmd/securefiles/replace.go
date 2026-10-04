package securefiles

import (
	"fmt"
	"os"
	"strings"

	"azuredevops/azdevops/organization"
	"azuredevops/azdevops/securefile"
	"azuredevops/cmd"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var replaceCmd = &cobra.Command{
	Use:   "replace <archivo-local>",
	Short: "Reemplaza el contenido de un archivo seguro conservando sus permisos",
	Long: `Reemplaza el contenido de un archivo seguro existente conservando:

  · el nombre y sus propiedades,
  · los pipelines autorizados (o el acceso abierto a todos los pipelines),
  · los roles asignados (quién puede usarlo, verlo o administrarlo) y si hereda los del proyecto,
  · las aprobaciones y checks configurados.

Azure DevOps no permite cambiar el contenido de un archivo seguro ni tener dos con el
mismo nombre, así que el proceso es: renombrar el actual, subir el nuevo con el nombre
original, copiar toda la configuración y eliminar el anterior. Si algo falla antes de
eliminarlo, los cambios se revierten.

Nota: el archivo nuevo tiene otro ID. Los pipelines YAML lo referencian por nombre, así
que siguen funcionando; si algún release clásico lo referencia por ID habría que revisarlo.`,
	Example: `  azdevops securefiles replace ./cert-2026.pfx --name cert.pfx
  azdevops sf replace ./kubeconfig --name kubeconfig-prod --dry-run
  azdevops sf replace ./cert.pfx --name cert.pfx --keep-old --yes`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		file, err := resolveFile(c, client, "Archivo seguro a reemplazar")
		if err != nil {
			return err
		}
		path, err := localPath(args, "Ruta del nuevo archivo")
		if err != nil {
			return err
		}
		data, err := os.ReadFile(strings.TrimSpace(path))
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return fmt.Errorf("%s está vacío", path)
		}

		project, err := organization.GetProject(client)
		if err != nil {
			return err
		}
		var snap *securefile.Snapshot
		if err := ui.Spinner("Leyendo la configuración actual...", func() (err error) {
			snap, err = securefile.TakeSnapshot(client, project.ID, *file)
			return err
		}); err != nil {
			return err
		}

		printPlan(snap, path, len(data))
		if dry, _ := c.Flags().GetBool("dry-run"); dry {
			ui.Info("--dry-run: no se hizo ningún cambio")
			return nil
		}
		yes, _ := c.Flags().GetBool("yes")
		if err := cmd.Confirm(fmt.Sprintf("¿Reemplazar el contenido de '%s'?", file.Name), yes); err != nil {
			return err
		}

		keepOld, _ := c.Flags().GetBool("keep-old")
		me, err := organization.CurrentUser(client)
		if err != nil {
			return err
		}
		progress := ui.StartProgress("Reemplazando...")
		res, err := securefile.Replace(client, project.ID, snap, data, securefile.ReplaceOptions{
			KeepOld:  keepOld,
			SelfID:   me.ID,
			Progress: progress.SetTitle,
		})
		progress.Stop()
		if err != nil {
			return err
		}

		ui.Success("'%s' reemplazado (nuevo ID %s)", res.New.Name, res.New.ID)
		if len(res.Pipelines) > 0 || res.AllPipelines {
			ui.Success("Pipelines autorizados conservados: %d%s", len(res.Pipelines), map[bool]string{true: " + acceso abierto", false: ""}[res.AllPipelines])
		}
		if res.Roles > 0 {
			ui.Success("Roles asignados conservados: %d", res.Roles)
		}
		if !res.Inherits {
			ui.Success("Herencia de permisos desactivada, igual que el original")
		}
		if res.Checks > 0 {
			ui.Success("Aprobaciones/checks conservados: %d", res.Checks)
		}
		for _, w := range res.Warnings {
			ui.Warn("%s", w)
		}
		switch {
		case res.DeleteOldError != "":
			ui.Warn("El nuevo archivo está listo, pero no se pudo eliminar el anterior ('%s'): %s", res.OldRenamedTo, res.DeleteOldError)
		case keepOld:
			ui.Info("El archivo anterior se conservó como '%s'", res.OldRenamedTo)
		}
		if cmd.WantsData() {
			return cmd.Print(res)
		}
		return nil
	},
}

func printPlan(snap *securefile.Snapshot, path string, size int) {
	fmt.Fprintf(ui.Err, "Se reemplazará %s con %s (%d bytes) conservando:\n", ui.Bold(snap.File.Name), path, size)
	ids := snap.Permissions.AuthorizedPipelineIDs()
	pipelines := fmt.Sprintf("%d pipeline(s) autorizados", len(ids))
	if len(ids) > 0 {
		strs := make([]string, len(ids))
		for i, id := range ids {
			strs[i] = fmt.Sprintf("#%d", id)
		}
		pipelines += " (" + strings.Join(strs, ", ") + ")"
	}
	if snap.Permissions.OpenToAllPipelines() {
		pipelines += " + acceso abierto a todos los pipelines"
	}
	fmt.Fprintf(ui.Err, "  · %s\n", pipelines)
	assigned := snap.AssignedRoles()
	roleNames := make([]string, len(assigned))
	for i, r := range assigned {
		roleNames[i] = fmt.Sprintf("%s (%s)", identity(&r.Identity), r.Role.Name)
	}
	fmt.Fprintf(ui.Err, "  · %d rol(es) asignado(s): %s\n", len(assigned), strings.Join(roleNames, ", "))
	if snap.Inherits() {
		fmt.Fprintln(ui.Err, "  · herencia de permisos del proyecto: activada")
	} else {
		fmt.Fprintln(ui.Err, "  · herencia de permisos del proyecto: desactivada")
	}
	fmt.Fprintf(ui.Err, "  · %d aprobación(es)/check(s)\n", len(snap.Checks))
	if len(snap.File.Properties) > 0 {
		fmt.Fprintf(ui.Err, "  · %d propiedad(es)\n", len(snap.File.Properties))
	}
}

func init() {
	replaceCmd.Flags().StringP("name", "n", "", "Nombre (o ID) del archivo seguro a reemplazar")
	replaceCmd.Flags().Bool("keep-old", false, "Conservar el archivo anterior renombrado en lugar de eliminarlo")
	replaceCmd.Flags().Bool("dry-run", false, "Mostrar qué se conservaría sin hacer cambios")
	replaceCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
	secureFilesCmd.AddCommand(replaceCmd)
}
