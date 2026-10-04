package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/azdevops/git"
	"azuredevops/azdevops/pipeline"
	"azuredevops/azdevops/serviceendpoint"
	"azuredevops/cmd"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var createPipelineCmd = &cobra.Command{
	Use:   "create",
	Short: "Crea un pipeline a partir de un archivo YAML de un repositorio",
	Example: `  azdevops pipelines create --name MiPipeline --repo-name mi-repo --yaml-path azure-pipelines.yml
  azdevops pipelines create -n MiPipeline -t gitHub -r org/repo -p .azure/pipeline.yml -s <id-conexion>
  azdevops pipelines create   # modo guiado`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		f := c.Flags()
		opts := pipeline.CreateOptions{}
		opts.Name, _ = f.GetString("name")
		opts.RepoType, _ = f.GetString("repo-type")
		opts.RepoName, _ = f.GetString("repo-name")
		opts.YAMLPath, _ = f.GetString("yaml-path")
		opts.Folder, _ = f.GetString("folder")
		opts.ServiceConnectionID, _ = f.GetString("service-connection")
		if f.Changed("branch") {
			opts.Branch, _ = f.GetString("branch")
		}

		if opts.Name == "" || opts.RepoName == "" || opts.YAMLPath == "" {
			if !ui.Interactive() {
				return ui.MissingError("name, --repo-name y --yaml-path")
			}
			if err := promptCreate(client, &opts, f.Changed("repo-type")); err != nil {
				return err
			}
		}

		created, err := pipeline.CreatePipeline(client, opts)
		if err != nil {
			if created == nil {
				return err
			}
			ui.Warn("%v", err)
		}
		ui.Success("Pipeline '%s' creado (ID: %d)", created.Name, created.ID)
		ui.Link(created.Links.Web.Href)
		if cmd.WantsJSON() {
			return ui.PrintJSON(created)
		}
		return nil
	},
}

func promptCreate(client *azdevops.Client, opts *pipeline.CreateOptions, repoTypeSet bool) error {
	var err error
	if opts.Name == "" {
		if opts.Name, err = ui.Input("Nombre del pipeline", "", "", ui.Required); err != nil {
			return err
		}
	}
	if !repoTypeSet {
		if opts.RepoType, err = ui.Select("Tipo de repositorio", []ui.Option[string]{
			{Label: "Azure Repos Git", Value: "azureReposGit"},
			{Label: "GitHub", Value: "gitHub"},
		}); err != nil {
			return err
		}
	}
	defaultBranch := "main"
	if opts.RepoName == "" {
		if opts.RepoType == "azureReposGit" {
			repos, _ := git.ListRepositories(client)
			if len(repos) > 0 {
				choices := make([]ui.Option[int], len(repos))
				for i, r := range repos {
					choices[i] = ui.Option[int]{Label: r.Name, Value: i}
				}
				idx, err := ui.Select("Repositorio", choices)
				if err != nil {
					return err
				}
				opts.RepoName = repos[idx].Name
				if b := cmd.ShortBranch(repos[idx].DefaultBranch); b != "" {
					defaultBranch = b
				}
			}
		}
		if opts.RepoName == "" {
			if opts.RepoName, err = ui.Input("Repositorio", "Para GitHub usa organización/repositorio", "", ui.Required); err != nil {
				return err
			}
		}
	}
	if opts.Branch == "" {
		if opts.Branch, err = ui.Input("Rama por defecto", "", defaultBranch, ui.Required); err != nil {
			return err
		}
	}
	if opts.YAMLPath == "" {
		if opts.YAMLPath, err = ui.Input("Ruta del archivo YAML", "Relativa a la raíz del repositorio", "azure-pipelines.yml", ui.Required); err != nil {
			return err
		}
	}
	if opts.Folder == "" || opts.Folder == "\\" {
		if opts.Folder, err = ui.Input("Carpeta", "Carpeta donde se organizará el pipeline", "\\", nil); err != nil {
			return err
		}
	}
	if opts.ServiceConnectionID == "" && opts.RepoType == "gitHub" {
		endpoints, _ := serviceendpoint.ListEndpoints(client, "github")
		if len(endpoints) > 0 {
			choices := []ui.Option[string]{{Label: "(ninguna)", Value: ""}}
			for _, e := range endpoints {
				choices = append(choices, ui.Option[string]{Label: e.Name, Value: e.ID})
			}
			if opts.ServiceConnectionID, err = ui.Select("Conexión de servicio de GitHub", choices); err != nil {
				return err
			}
		}
	}
	return nil
}

func init() {
	f := createPipelineCmd.Flags()
	f.StringP("name", "n", "", "Nombre del pipeline (obligatorio)")
	f.StringP("repo-type", "t", "azureReposGit", "Tipo de repositorio ('azureReposGit' o 'gitHub')")
	f.StringP("repo-name", "r", "", "Nombre del repositorio. Para GitHub, 'organizacion/repositorio' (obligatorio)")
	f.StringP("branch", "b", "main", "Rama por defecto del pipeline")
	f.StringP("yaml-path", "p", "", "Ruta al archivo YAML del pipeline (obligatorio)")
	f.String("folder", "\\", "Carpeta para organizar el pipeline")
	f.StringP("service-connection", "s", "", "ID de la conexión de servicio (repos de GitHub)")
	_ = createPipelineCmd.RegisterFlagCompletionFunc("repo-type", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"azureReposGit", "gitHub"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Pipelines.AddCommand(createPipelineCmd)
}
