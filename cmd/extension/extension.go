// Package extension contiene los comandos para crear y mantener extensiones de Azure DevOps.
package extension

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"azuredevops/cmd"
	"azuredevops/internal/scaffold"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var extensionCmd = &cobra.Command{
	Use:     "extension",
	Aliases: []string{"ext"},
	Short:   "Crea y mantiene extensiones de Azure DevOps (tareas de pipeline)",
	Long: `Crea y mantiene extensiones de Azure DevOps con la estructura:

  src/tasks/<tarea>/     código TypeScript de cada tarea (azure-pipelines-task-lib, jest)
  tasks/<tarea>/         salida empaquetada con ncc en un único index.js
  config/dev.json        overrides para el paquete privado de desarrollo
  config/release.json    overrides para el paquete público
  images/                iconos de la extensión y de las conexiones de servicio

azde-scripts instala y compila todas las tareas y tfx-cli genera el .vsix.`,
}

var uuidCmd = &cobra.Command{
	Use:   "uuid",
	Short: "Genera UUIDs (por ejemplo para el id de una tarea)",
	RunE: func(c *cobra.Command, args []string) error {
		n, _ := c.Flags().GetInt("count")
		for i := 0; i < n; i++ {
			fmt.Fprintln(ui.Out, scaffold.NewUUID())
		}
		return nil
	},
}

var resetIDsCmd = &cobra.Command{
	Use:   "reset-ids",
	Short: "Genera UUIDs nuevos para todas las tareas (al partir de otra extensión)",
	Long: `Asigna un id nuevo a cada task.json. Úsalo cuando copias una extensión existente
como base: dos tareas con el mismo id en la organización se pisan entre sí.`,
	RunE: func(c *cobra.Command, args []string) error {
		root, err := projectRoot(c)
		if err != nil {
			return err
		}
		yes, _ := c.Flags().GetBool("yes")
		if err := cmd.Confirm("¿Generar ids nuevos para todas las tareas? Las tareas instaladas con el id anterior dejarán de actualizarse", yes); err != nil {
			return err
		}
		changes, err := scaffold.ResetIDs(root)
		if err != nil {
			return err
		}
		for _, ch := range changes {
			ui.Success("%s: %s → %s", ch.File, ch.From, ch.To)
		}
		return nil
	},
}

var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Incrementa la versión de las tareas y de la extensión",
	Long: `Incrementa la versión de las tareas (task.json) y de la extensión (config/*.json).

Azure DevOps solo actualiza una tarea en los agentes si su versión cambia, así que
cada publicación necesita subir la versión de las tareas modificadas y la de la
extensión. Por defecto sube el patch de todas las tareas y de config/dev.json.`,
	Example: `  azdevops ext bump                         # patch: todas las tareas + dev
  azdevops ext bump --level minor --release  # minor: tareas + release
  azdevops ext bump --task mi-tarea --no-config
  azdevops ext bump --dry-run`,
	RunE: func(c *cobra.Command, args []string) error {
		root, err := projectRoot(c)
		if err != nil {
			return err
		}
		f := c.Flags()
		level, _ := f.GetString("level")
		if level != scaffold.LevelMajor && level != scaffold.LevelMinor && level != scaffold.LevelPatch {
			return fmt.Errorf("--level debe ser major, minor o patch")
		}
		only, _ := f.GetStringSlice("task")
		dry, _ := f.GetBool("dry-run")
		noTasks, _ := f.GetBool("no-tasks")
		noConfig, _ := f.GetBool("no-config")
		release, _ := f.GetBool("release")
		all, _ := f.GetBool("all-configs")
		configs := []string{"dev"}
		switch {
		case all:
			configs = []string{"dev", "release"}
		case release:
			configs = []string{"release"}
		}

		var changes []scaffold.Change
		if !noTasks {
			ch, err := scaffold.BumpTasks(root, level, only, dry)
			if err != nil {
				return err
			}
			changes = append(changes, ch...)
		}
		if !noConfig {
			ch, err := scaffold.BumpConfigs(root, level, configs, dry)
			if err != nil {
				return err
			}
			changes = append(changes, ch...)
		}
		if cmd.WantsData() {
			return cmd.Print(changes)
		}
		rows := make([][]string, len(changes))
		for i, ch := range changes {
			rows[i] = []string{ch.File, ch.From, ch.To}
		}
		ui.Table([]string{"archivo", "antes", "después"}, rows)
		if dry {
			ui.Info("--dry-run: no se modificó ningún archivo")
		}
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Revisa la extensión antes de empaquetarla (ids, contribuciones, iconos, versiones)",
	RunE: func(c *cobra.Command, args []string) error {
		root, err := projectRoot(c)
		if err != nil {
			return err
		}
		issues := scaffold.Validate(root)
		if cmd.WantsData() {
			if issues == nil {
				issues = []scaffold.Issue{}
			}
			return cmd.Print(issues)
		}
		errorsFound := 0
		for _, i := range issues {
			if i.Severity == "error" {
				errorsFound++
				ui.Errorf("%s: %s", i.File, i.Message)
			} else {
				ui.Warn("%s: %s", i.File, i.Message)
			}
		}
		if errorsFound > 0 {
			return fmt.Errorf("%d error(es) en la extensión", errorsFound)
		}
		ui.Success("La extensión en %s es válida", root)
		return nil
	},
}

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Empaqueta las tareas con ncc y genera el .vsix (dev por defecto)",
	Example: `  azdevops ext pack
  azdevops ext pack --release --rev-version`,
	RunE: func(c *cobra.Command, args []string) error {
		root, err := projectRoot(c)
		if err != nil {
			return err
		}
		if err := checkReady(root); err != nil {
			return err
		}
		release, _ := c.Flags().GetBool("release")
		rev, _ := c.Flags().GetBool("rev-version")
		script := "pack"
		if rev {
			script = "packupversion"
		}
		if !release {
			script += ":dev"
		}
		return run(root, "npm", "run", script)
	},
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publica la extensión en el Marketplace (dev por defecto) y la comparte",
	Long: `Empaqueta las tareas y publica con tfx-cli usando config/dev.json (o
config/release.json con --release).

Necesita un PAT con el scope "Marketplace (Publish)" y "All accessible
organizations": --token, la variable AZURE_MARKETPLACE_TOKEN o se pedirá.
La versión de desarrollo se comparte por defecto con la organización de tu perfil.`,
	Example: `  azdevops ext publish --rev-version
  azdevops ext publish --share-with otra-org
  azdevops ext publish --release`,
	RunE: func(c *cobra.Command, args []string) error {
		root, err := projectRoot(c)
		if err != nil {
			return err
		}
		if err := checkReady(root); err != nil {
			return err
		}
		if issues := scaffold.Validate(root); hasErrors(issues) {
			return fmt.Errorf("la extensión tiene errores; ejecuta 'azdevops ext validate'")
		}
		f := c.Flags()
		release, _ := f.GetBool("release")
		rev, _ := f.GetBool("rev-version")
		shareWith, _ := f.GetStringSlice("share-with")
		token, _ := f.GetString("token")
		if token == "" {
			token = os.Getenv("AZURE_MARKETPLACE_TOKEN")
		}
		if token == "" {
			if !ui.Interactive() {
				return fmt.Errorf("falta el token del Marketplace: usa --token o AZURE_MARKETPLACE_TOKEN")
			}
			if token, err = ui.Password("PAT del Marketplace", "Scope: Marketplace (Publish), All accessible organizations"); err != nil {
				return err
			}
		}
		if len(shareWith) == 0 && !release {
			if settings, _, err := cmd.ResolveSettings(); err == nil && settings.Org != "" {
				shareWith = []string{settings.Org}
			}
		}
		config := "config/dev.json"
		if release {
			config = "config/release.json"
		}

		if err := run(root, "npm", "run", "package:tasks"); err != nil {
			return err
		}
		tfxArgs := []string{"--no-install", "tfx", "extension", "publish", "--manifest-globs", "vss-extension.json",
			"--overrides-file", config, "--token", token, "--no-prompt"}
		if rev {
			tfxArgs = append(tfxArgs, "--rev-version")
		}
		if len(shareWith) > 0 {
			tfxArgs = append(tfxArgs, "--share-with")
			tfxArgs = append(tfxArgs, shareWith...)
		}
		ui.Info("Publicando con %s%s", config, sharedSuffix(shareWith))
		return run(root, "npx", tfxArgs...)
	},
}

func sharedSuffix(orgs []string) string {
	if len(orgs) == 0 {
		return ""
	}
	return " y compartiendo con " + strings.Join(orgs, ", ")
}

func hasErrors(issues []scaffold.Issue) bool {
	for _, i := range issues {
		if i.Severity == "error" {
			return true
		}
	}
	return false
}

// projectRoot localiza la raíz de la extensión desde --dir o el directorio actual.
func projectRoot(c *cobra.Command) (string, error) {
	dir, _ := c.Flags().GetString("dir")
	if dir == "" {
		dir = "."
	}
	return scaffold.FindRoot(dir)
}

func checkReady(root string) error {
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("se necesita Node.js y npm en el PATH")
	}
	if _, err := os.Stat(filepath.Join(root, "node_modules")); err != nil {
		return fmt.Errorf("faltan dependencias: ejecuta 'npm install' en %s", root)
	}
	return nil
}

// run ejecuta un comando en dir mostrando su salida.
func run(dir, name string, args ...string) error {
	shown := make([]string, len(args))
	for i, a := range args {
		shown[i] = a
		if i > 0 && args[i-1] == "--token" {
			shown[i] = "****"
		}
	}
	ui.Info("%s %s", name, strings.Join(shown, " "))
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Stdout, c.Stderr, c.Stdin = os.Stderr, os.Stderr, os.Stdin
	if err := c.Run(); err != nil {
		label := name
		for _, a := range shown {
			if !strings.HasPrefix(a, "-") {
				label += " " + a
				if len(strings.Fields(label)) >= 3 {
					break
				}
			}
		}
		return fmt.Errorf("'%s' falló: %w", label, err)
	}
	return nil
}

func init() {
	uuidCmd.Flags().IntP("count", "n", 1, "Cantidad de UUIDs")
	resetIDsCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")

	bf := bumpCmd.Flags()
	bf.String("level", scaffold.LevelPatch, "Nivel: major, minor o patch")
	bf.StringSlice("task", nil, "Solo estas tareas (carpeta en src/tasks; se puede repetir)")
	bf.Bool("release", false, "Subir config/release.json en lugar de config/dev.json")
	bf.Bool("all-configs", false, "Subir config/dev.json y config/release.json")
	bf.Bool("no-tasks", false, "No tocar las tareas")
	bf.Bool("no-config", false, "No tocar config/*.json")
	bf.Bool("dry-run", false, "Mostrar los cambios sin aplicarlos")

	packCmd.Flags().Bool("release", false, "Usar config/release.json")
	packCmd.Flags().Bool("rev-version", false, "Incrementar la versión de la extensión al empaquetar")
	publishCmd.Flags().Bool("release", false, "Usar config/release.json")
	publishCmd.Flags().Bool("rev-version", false, "Incrementar la versión de la extensión al publicar")
	publishCmd.Flags().StringSlice("share-with", nil, "Organizaciones con las que compartir (por defecto la de tu perfil en dev)")
	publishCmd.Flags().String("token", "", "PAT del Marketplace (env: AZURE_MARKETPLACE_TOKEN)")

	for _, c := range []*cobra.Command{resetIDsCmd, bumpCmd, validateCmd, packCmd, publishCmd} {
		c.Flags().String("dir", "", "Carpeta de la extensión (por defecto se busca desde la actual)")
	}
	extensionCmd.AddCommand(uuidCmd, resetIDsCmd, bumpCmd, validateCmd, packCmd, publishCmd)
	cmd.RootCmd.AddCommand(extensionCmd)
}
