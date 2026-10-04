package cmd

import (
	"fmt"

	"azuredevops/azdevops"
	"azuredevops/azdevops/organization"
	"azuredevops/internal/config"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Guarda organización, proyecto y PAT en un perfil",
	Long: `Valida las credenciales contra Azure DevOps y las guarda en un perfil
(~/.config/azdevops/config.json, permisos 0600).

En una terminal se te preguntará lo que falte; para scripts usa:
  azdevops login --org mi-org --project mi-proyecto --pat $AZURE_PAT [--profile trabajo]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := globalFlags.profile
		if name == "" {
			name = config.DefaultProfile
		}
		defaults := config.Profile{Org: globalFlags.org, Project: globalFlags.project, PAT: globalFlags.pat}
		if defaults.Org != "" && defaults.Project != "" && defaults.PAT != "" {
			return saveProfile(name, defaults)
		}
		if !ui.Interactive() {
			return fmt.Errorf("en modo no interactivo debes indicar --org, --project y --pat")
		}
		_, err := runLoginWizard(name, defaults)
		return err
	},
}

func runLoginWizard(name string, defaults config.Profile) (config.Profile, error) {
	p := defaults
	var err error
	if p.Org, err = ui.Input("Organización", "El nombre que aparece en https://dev.azure.com/<organización>", p.Org, ui.Required); err != nil {
		return p, err
	}
	if p.PAT == "" {
		if p.PAT, err = ui.Password("Personal Access Token (PAT)", "Necesita permisos sobre Build, Variable Groups y Graph según lo que vayas a usar"); err != nil {
			return p, err
		}
	}

	if p.Project == "" {
		var projects []string
		_ = ui.Spinner("Consultando proyectos...", func() error {
			list, err := organization.ListProjects(azdevops.NewClient(p.Org, "", p.PAT))
			for _, pr := range list {
				projects = append(projects, pr.Name)
			}
			return err
		})
		if len(projects) > 0 {
			opts := make([]ui.Option[string], len(projects))
			for i, pr := range projects {
				opts[i] = ui.Option[string]{Label: pr, Value: pr}
			}
			if p.Project, err = ui.Select("Proyecto", opts); err != nil {
				return p, err
			}
		} else if p.Project, err = ui.Input("Proyecto", "", "", ui.Required); err != nil {
			return p, err
		}
	}
	return p, saveProfile(name, p)
}

func saveProfile(name string, p config.Profile) error {
	client := azdevops.NewClient(p.Org, p.Project, p.PAT)
	client.Debug = globalFlags.debug
	if err := ui.Spinner("Validando credenciales...", func() error {
		_, err := organization.GetProject(client)
		return err
	}); err != nil {
		return fmt.Errorf("no se pudieron validar las credenciales: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Profiles[name] = p
	if cfg.Current == "" || len(cfg.Profiles) == 1 {
		cfg.Current = name
	}
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("no se pudo guardar la configuración: %w", err)
	}
	path, _ := config.Path()
	ui.Success("Perfil '%s' guardado (%s/%s) en %s", name, p.Org, p.Project, path)
	if cfg.Current != name {
		ui.Info("Para usarlo por defecto: azdevops config use %s", name)
	}
	ResetClient()
	return nil
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
