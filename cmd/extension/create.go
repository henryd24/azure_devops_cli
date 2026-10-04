package extension

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"azuredevops/internal/scaffold"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [carpeta]",
	Short: "Crea una extensión nueva lista para compilar, probar y empaquetar",
	Long: `Crea una extensión con una o más tareas, ids (UUID) nuevos, configuraciones
dev/release, iconos de ejemplo, tests con jest, empaquetado con ncc y, opcionalmente,
un tipo de conexión de servicio propio.

En una terminal se te preguntará todo lo que no indiques con flags.`,
	Example: `  azdevops ext init
  azdevops ext init mi-extension --name "Mi Extensión" --publisher hendamm --task hola-mundo --yes
  azdevops ext init deploy-tools --publisher hendamm --task deploy --task rollback --endpoint --install --git`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		f := c.Flags()
		yes, _ := f.GetBool("yes")
		guided := ui.Interactive() && !yes

		p := scaffold.Project{}
		p.Name, _ = f.GetString("name")
		p.ID, _ = f.GetString("id")
		p.Description, _ = f.GetString("description")
		p.Publisher, _ = f.GetString("publisher")
		p.DevPublisher, _ = f.GetString("dev-publisher")
		p.Author, _ = f.GetString("author")
		p.Repository, _ = f.GetString("repo")
		p.Category, _ = f.GetString("category")
		p.Public, _ = f.GetBool("public")
		p.Version, _ = f.GetString("version")
		license, _ := f.GetString("license")
		taskSlugs, _ := f.GetStringSlice("task")
		taskCategory, _ := f.GetString("task-category")
		withEndpoint, _ := f.GetBool("endpoint")
		endpointName, _ := f.GetString("endpoint-name")
		endpointURL, _ := f.GetString("endpoint-url")
		install, _ := f.GetBool("install")
		gitInit, _ := f.GetBool("git")

		dir := ""
		if len(args) == 1 {
			dir = args[0]
		}
		if p.Name == "" && dir != "" {
			p.Name = scaffold.Title(scaffold.Slugify(filepath.Base(dir)))
		}
		if p.Author == "" {
			p.Author = gitConfig("user.name")
		}

		var err error
		ask := func(dst *string, title, desc string, validate func(string) error) error {
			if !guided {
				return nil
			}
			v, err := ui.Input(title, desc, *dst, validate)
			if err == nil {
				*dst = v
			}
			return err
		}

		if err = ask(&p.Name, "Nombre de la extensión", "Como se verá en el Marketplace", ui.Required); err != nil {
			return err
		}
		if p.Name == "" {
			return ui.MissingError("name")
		}
		if p.ID == "" {
			p.ID = scaffold.Slugify(p.Name)
		}
		if err = ask(&p.ID, "ID de la extensión", "kebab-case, único para tu publisher", validSlug); err != nil {
			return err
		}
		if err = ask(&p.Description, "Descripción", "", nil); err != nil {
			return err
		}
		if err = ask(&p.Publisher, "Publisher (release)", "ID de tu publisher en el Marketplace", validSlug); err != nil {
			return err
		}
		if p.Publisher == "" {
			return ui.MissingError("publisher")
		}
		if p.DevPublisher == "" {
			p.DevPublisher = "dev-" + p.Publisher
		}
		if err = ask(&p.DevPublisher, "Publisher (dev)", "Publisher para la versión privada de pruebas", validSlug); err != nil {
			return err
		}
		if err = ask(&p.Author, "Autor", "", nil); err != nil {
			return err
		}
		if p.Repository == "" && guided {
			p.Repository = repoGuess(p.ID)
		}
		if err = ask(&p.Repository, "Repositorio (opcional)", "URL del repositorio git", nil); err != nil {
			return err
		}
		if guided {
			if !f.Changed("category") {
				if p.Category, err = ui.Select("Categoría del Marketplace", options(scaffold.ExtensionCategories)); err != nil {
					return err
				}
			}
			if !f.Changed("public") {
				if p.Public, err = ui.Confirm("¿La versión release será pública en el Marketplace?", true); err != nil {
					return err
				}
			}
			if err = ask(&p.Version, "Versión inicial", "", validSemver); err != nil {
				return err
			}
			if !f.Changed("license") {
				if license, err = ui.Select("Licencia", []ui.Option[string]{{Label: "MIT", Value: "MIT"}, {Label: "Sin licencia (privada)", Value: "UNLICENSED"}}); err != nil {
					return err
				}
			}
		}
		if p.Description == "" {
			p.Description = p.Name
		}
		if err := validSlug(p.ID); err != nil {
			return fmt.Errorf("--id: %w", err)
		}
		if err := validSemver(p.Version); err != nil {
			return fmt.Errorf("--version: %w", err)
		}
		p.LicenseID = strings.ToUpper(license)
		if p.LicenseID != "MIT" {
			p.LicenseID = "UNLICENSED"
		}

		var endpoint *scaffold.Endpoint
		if guided && !f.Changed("endpoint") {
			if withEndpoint, err = ui.Confirm("¿Incluir un tipo de conexión de servicio propio (URL + token)?", false); err != nil {
				return err
			}
		}
		if withEndpoint {
			if endpointName == "" {
				endpointName = p.Name
			}
			if err = ask(&endpointName, "Nombre visible de la conexión de servicio", "", ui.Required); err != nil {
				return err
			}
			if err = ask(&endpointURL, "URL por defecto del servicio", "", nil); err != nil {
				return err
			}
			endpoint = scaffold.NewEndpoint(p.ID, endpointName, endpointURL)
		}

		var tasks []scaffold.Task
		for _, slug := range taskSlugs {
			if err := validSlug(slug); err != nil {
				return fmt.Errorf("--task %s: %w", slug, err)
			}
			tasks = append(tasks, scaffold.NewTask(p.Publisher, slug, "", "", taskCategory, endpoint))
		}
		if guided && len(tasks) == 0 {
			for {
				t, err := promptTask(p, endpoint, len(tasks) == 0, defaultTaskSlug(p.ID, len(tasks)))
				if err != nil {
					return err
				}
				tasks = append(tasks, t)
				more, err := ui.Confirm("¿Agregar otra tarea?", false)
				if err != nil {
					return err
				}
				if !more {
					break
				}
			}
		}
		if len(tasks) == 0 {
			tasks = append(tasks, scaffold.NewTask(p.Publisher, defaultTaskSlug(p.ID, 0), "", "", taskCategory, endpoint))
		}

		if dir == "" {
			dir = p.ID
		}
		if guided {
			if !f.Changed("install") {
				if install, err = ui.Confirm("¿Ejecutar npm install al terminar?", true); err != nil {
					return err
				}
			}
			if !f.Changed("git") {
				if gitInit, err = ui.Confirm("¿Inicializar un repositorio git?", true); err != nil {
					return err
				}
			}
		}

		var endpoints []scaffold.Endpoint
		if endpoint != nil {
			endpoints = []scaffold.Endpoint{*endpoint}
		}
		created, err := scaffold.Generate(dir, p, tasks, endpoints)
		if err != nil {
			return err
		}
		ui.Success("Extensión '%s' creada en %s (%d archivos)", p.Name, dir, len(created))
		rows := make([][]string, len(tasks))
		for i, t := range tasks {
			rows[i] = []string{t.Slug, t.Name, t.UUID}
		}
		ui.Table([]string{"tarea", "name (task.json)", "id"}, rows)
		if endpoint != nil {
			ui.Info("Conexión de servicio: %s (connectedService:%s)", endpoint.DisplayName, endpoint.Name)
		}

		if gitInit {
			if err := run(dir, "git", "init", "-q"); err != nil {
				ui.Warn("%v", err)
			}
		}
		if install {
			if err := run(dir, "npm", "install"); err != nil {
				return err
			}
		}
		printNextSteps(dir, install)
		return nil
	},
}

var addTaskCmd = &cobra.Command{
	Use:   "add-task [nombre]",
	Short: "Agrega una tarea nueva a una extensión existente",
	Example: `  azdevops ext add-task deploy-app
  azdevops ext add-task rollback --friendly-name "Rollback App" --category Deploy --endpoint`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		root, err := projectRoot(c)
		if err != nil {
			return err
		}
		p, endpoints, err := scaffold.ProjectInfo(root)
		if err != nil {
			return err
		}
		f := c.Flags()
		friendly, _ := f.GetString("friendly-name")
		description, _ := f.GetString("description")
		category, _ := f.GetString("category")
		useEndpoint, _ := f.GetBool("endpoint")

		slug := ""
		if len(args) == 1 {
			slug = args[0]
		}
		var t scaffold.Task
		if slug == "" {
			if !ui.Interactive() {
				return fmt.Errorf("indica el nombre de la tarea")
			}
			var ep *scaffold.Endpoint
			if len(endpoints) > 0 {
				use, err := ui.Confirm(fmt.Sprintf("¿La tarea usará la conexión de servicio '%s'?", endpoints[0].DisplayName), false)
				if err != nil {
					return err
				}
				if use {
					ep = &endpoints[0]
				}
			}
			if t, err = promptTask(p, ep, false, ""); err != nil {
				return err
			}
		} else {
			if err := validSlug(slug); err != nil {
				return err
			}
			var ep *scaffold.Endpoint
			if useEndpoint {
				if len(endpoints) == 0 {
					return fmt.Errorf("la extensión no define ninguna conexión de servicio")
				}
				ep = &endpoints[0]
			}
			t = scaffold.NewTask(p.Publisher, slug, friendly, description, category, ep)
		}

		created, err := scaffold.AddTask(root, p, t)
		if err != nil {
			return err
		}
		for _, f := range created {
			fmt.Fprintf(ui.Err, "  + %s\n", f)
		}
		ui.Success("Tarea '%s' creada (name: %s, id: %s)", t.Slug, t.Name, t.UUID)
		ui.Info("Instala sus dependencias con: npm install (en %s)", root)
		return nil
	},
}

func promptTask(p scaffold.Project, endpoint *scaffold.Endpoint, first bool, def string) (scaffold.Task, error) {
	title := "Nombre de la tarea"
	if first {
		title = "Nombre de la primera tarea"
	}
	slug, err := ui.Input(title, "kebab-case; será la carpeta en src/tasks", def, validSlug)
	if err != nil {
		return scaffold.Task{}, err
	}
	friendly, err := ui.Input("Nombre visible de la tarea", "Como aparece en el catálogo de tareas", scaffold.Title(slug), ui.Required)
	if err != nil {
		return scaffold.Task{}, err
	}
	description, err := ui.Input("Descripción de la tarea", "", friendly, nil)
	if err != nil {
		return scaffold.Task{}, err
	}
	category, err := ui.Select("Categoría de la tarea", options(scaffold.TaskCategories))
	if err != nil {
		return scaffold.Task{}, err
	}
	return scaffold.NewTask(p.Publisher, slug, friendly, description, category, endpoint), nil
}

func printNextSteps(dir string, installed bool) {
	fmt.Fprintf(ui.Err, "\nSiguientes pasos:\n  cd %s\n", dir)
	if !installed {
		fmt.Fprintln(ui.Err, "  npm install")
	}
	fmt.Fprintln(ui.Err, "  npm test                       # tests de todas las tareas")
	fmt.Fprintln(ui.Err, "  azdevops ext pack              # genera el .vsix de dev")
	fmt.Fprintln(ui.Err, "  azdevops ext publish           # publica y comparte con tu organización")
	fmt.Fprintln(ui.Err, "  azdevops ext bump              # antes de cada publicación")
	fmt.Fprintln(ui.Err, "Reemplaza los iconos de images/ y de cada tarea (icon.png) por los definitivos.")
}

func defaultTaskSlug(id string, n int) string {
	if n == 0 {
		return id
	}
	return fmt.Sprintf("%s-%d", id, n+1)
}

func options(values []string) []ui.Option[string] {
	out := make([]ui.Option[string], len(values))
	for i, v := range values {
		out[i] = ui.Option[string]{Label: v, Value: v}
	}
	return out
}

func validSlug(s string) error {
	if !scaffold.IsSlug(s) {
		if suggestion := scaffold.Slugify(s); suggestion != "" {
			return fmt.Errorf("usa solo minúsculas, números y guiones (p. ej. %s)", suggestion)
		}
		return fmt.Errorf("usa solo minúsculas, números y guiones")
	}
	return nil
}

func validSemver(s string) error {
	if !scaffold.IsSemver(s) {
		return fmt.Errorf("usa el formato X.Y.Z")
	}
	return nil
}

func gitConfig(key string) string {
	out, err := exec.Command("git", "config", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func repoGuess(id string) string {
	user := gitConfig("github.user")
	if user == "" {
		return ""
	}
	return "https://github.com/" + user + "/" + id
}

func init() {
	f := initCmd.Flags()
	f.String("name", "", "Nombre de la extensión")
	f.String("id", "", "ID de la extensión (kebab-case; por defecto derivado del nombre)")
	f.String("description", "", "Descripción")
	f.String("publisher", "", "Publisher del Marketplace (release)")
	f.String("dev-publisher", "", "Publisher para la versión dev (por defecto dev-<publisher>)")
	f.String("author", "", "Autor (por defecto git config user.name)")
	f.String("repo", "", "URL del repositorio")
	f.String("category", "Azure Pipelines", "Categoría del Marketplace")
	f.Bool("public", true, "La versión release es pública")
	f.String("version", "0.1.0", "Versión inicial de la extensión")
	f.String("license", "MIT", "Licencia: MIT o none")
	f.StringSlice("task", nil, "Tarea a crear (kebab-case; se puede repetir)")
	f.String("task-category", "Utility", "Categoría de las tareas creadas con --task")
	f.Bool("endpoint", false, "Incluir un tipo de conexión de servicio propio")
	f.String("endpoint-name", "", "Nombre visible de la conexión de servicio")
	f.String("endpoint-url", "", "URL por defecto de la conexión de servicio")
	f.Bool("install", false, "Ejecutar npm install al terminar")
	f.Bool("git", false, "Ejecutar git init")
	f.BoolP("yes", "y", false, "No preguntar: usar flags y valores por defecto")
	_ = initCmd.RegisterFlagCompletionFunc("category", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return scaffold.ExtensionCategories, cobra.ShellCompDirectiveNoFileComp
	})

	af := addTaskCmd.Flags()
	af.String("friendly-name", "", "Nombre visible de la tarea")
	af.String("description", "", "Descripción de la tarea")
	af.String("category", "Utility", "Categoría de la tarea")
	af.Bool("endpoint", false, "Usar la conexión de servicio de la extensión")
	af.String("dir", "", "Carpeta de la extensión (por defecto se busca desde la actual)")
	_ = addTaskCmd.RegisterFlagCompletionFunc("category", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return scaffold.TaskCategories, cobra.ShellCompDirectiveNoFileComp
	})

	extensionCmd.AddCommand(initCmd, addTaskCmd)
}
