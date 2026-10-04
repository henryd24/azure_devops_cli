package cmd

import (
	"fmt"
	"net/url"
	"strconv"

	"azuredevops/azdevops"
	"azuredevops/azdevops/pipeline"
	vg "azuredevops/azdevops/variable_group"
	"azuredevops/internal/ui"

	"github.com/spf13/cobra"
)

var openTargets = []string{"project", "pipeline", "run", "variables", "workitem", "repo", "approvals", "environments", "service-connections", "securefiles", "boards"}

var openCmd = &cobra.Command{
	Use:   "open [recurso] [id|nombre]",
	Short: "Abre en el navegador el proyecto o un recurso",
	Long: `Abre en el navegador el portal de Azure DevOps.

Recursos: project (por defecto), pipeline [id|nombre], run <build-id>,
variables [nombre], workitem <id>, repo [nombre], approvals, environments,
service-connections, securefiles y boards.`,
	Example: `  azdevops open
  azdevops open pipeline 123
  azdevops open run 4567
  azdevops open variables app-dev
  azdevops open workitem 42 --print`,
	Args:      cobra.MaximumNArgs(2),
	ValidArgs: openTargets,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := NewClient()
		if err != nil {
			return err
		}
		target := "project"
		if len(args) > 0 {
			target = args[0]
		} else if ui.Interactive() {
			opts := make([]ui.Option[string], len(openTargets))
			for i, t := range openTargets {
				opts[i] = ui.Option[string]{Label: t, Value: t}
			}
			if target, err = ui.Select("¿Qué quieres abrir?", opts); err != nil {
				return err
			}
		}
		ref := ""
		if len(args) > 1 {
			ref = args[1]
		}

		link, err := resourceURL(client, target, ref)
		if err != nil {
			return err
		}
		if printOnly, _ := c.Flags().GetBool("print"); printOnly {
			fmt.Fprintln(ui.Out, link)
			return nil
		}
		if err := ui.OpenBrowser(link); err != nil {
			ui.Warn("No se pudo abrir el navegador (%v); abre este enlace:", err)
			fmt.Fprintln(ui.Out, link)
			return nil
		}
		ui.Success("Abriendo %s", link)
		return nil
	},
}

func resourceURL(client *azdevops.Client, target, ref string) (string, error) {
	needID := func(title string, pick func() (int, error)) (int, error) {
		if ref != "" {
			id, err := strconv.Atoi(ref)
			if err != nil {
				return 0, fmt.Errorf("%s: '%s' no es un ID válido", title, ref)
			}
			return id, nil
		}
		if pick == nil || !ui.Interactive() {
			return 0, fmt.Errorf("indica el ID: azdevops open %s <id>", target)
		}
		return pick()
	}

	switch target {
	case "project":
		return client.WebURL(""), nil
	case "boards":
		return client.WebURL("_boards"), nil
	case "approvals", "environments":
		return client.WebURL("_environments"), nil
	case "securefiles", "sf":
		return client.WebURL("_library?itemType=SecureFiles"), nil
	case "service-connections", "sc":
		return client.WebURL("_settings/adminservices"), nil
	case "pipeline", "pipelines":
		if ref == "" && !ui.Interactive() {
			return client.WebURL("_build"), nil
		}
		id, err := strconv.Atoi(ref)
		if ref != "" && err != nil {
			defs, err := pipeline.GetBuildDefinitionByName(client, ref)
			if err != nil {
				return "", err
			}
			if len(defs) == 0 || defs[0].ID == nil {
				return "", fmt.Errorf("no se encontró el pipeline '%s'", ref)
			}
			id = int(*defs[0].ID)
		} else if ref == "" {
			if id, err = PickPipeline(client, "Pipeline"); err != nil {
				return "", err
			}
		}
		return client.WebURL(fmt.Sprintf("_build?definitionId=%d", id)), nil
	case "run", "build":
		id, err := needID("run", func() (int, error) { return PickBuild(client, 0, "Ejecución") })
		if err != nil {
			return "", err
		}
		return client.WebURL(fmt.Sprintf("_build/results?buildId=%d", id)), nil
	case "variables", "vg":
		if ref == "" && !ui.Interactive() {
			return client.WebURL("_library?itemType=VariableGroups"), nil
		}
		name := ref
		if name == "" {
			var err error
			if name, err = PickVariableGroupName(client, "Variable Group"); err != nil {
				return "", err
			}
		}
		groups, err := vg.GetVariableGroupByName(client, name)
		if err != nil {
			return "", err
		}
		if len(groups) == 0 {
			return "", fmt.Errorf("no se encontró el Variable Group '%s'", name)
		}
		return client.VariableGroupWebURL(groups[0].Id), nil
	case "workitem", "wi":
		id, err := needID("workitem", func() (int, error) {
			v, err := ui.Input("ID del work item", "", "", PositiveInt)
			if err != nil {
				return 0, err
			}
			return strconv.Atoi(v)
		})
		if err != nil {
			return "", err
		}
		return client.WebURL(fmt.Sprintf("_workitems/edit/%d", id)), nil
	case "repo", "repos":
		if ref == "" {
			return client.WebURL("_git/" + url.PathEscape(client.Project)), nil
		}
		return client.WebURL("_git/" + url.PathEscape(ref)), nil
	}
	return "", fmt.Errorf("recurso '%s' desconocido (opciones: %v)", target, openTargets)
}

func init() {
	openCmd.Flags().Bool("print", false, "Solo imprimir la URL, sin abrir el navegador")
	RootCmd.AddCommand(openCmd)
}
