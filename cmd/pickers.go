package cmd

import (
	"fmt"
	"strconv"

	"azuredevops/azdevops"
	"azuredevops/azdevops/pipeline"
	vg "azuredevops/azdevops/variable_group"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

func variableGroupOptions(client *azdevops.Client) ([]ui.Option[string], error) {
	var groups []models.VariableGroup
	err := ui.Spinner("Cargando Variable Groups...", func() (err error) {
		groups, err = vg.ListVariableGroups(client, "")
		return err
	})
	if err != nil {
		return nil, err
	}
	opts := make([]ui.Option[string], len(groups))
	for i, g := range groups {
		opts[i] = ui.Option[string]{Label: fmt.Sprintf("%s  (%d variables)", g.Name, len(g.Variables)), Value: g.Name}
	}
	return opts, nil
}

// PickVariableGroupName permite elegir un Variable Group de la lista.
func PickVariableGroupName(client *azdevops.Client, title string) (string, error) {
	opts, err := variableGroupOptions(client)
	if err != nil {
		return "", err
	}
	return ui.Select(title, opts)
}

// PickVariableGroupNames permite elegir varios Variable Groups.
func PickVariableGroupNames(client *azdevops.Client, title string) ([]string, error) {
	opts, err := variableGroupOptions(client)
	if err != nil {
		return nil, err
	}
	return ui.MultiSelect(title, opts)
}

// PickPipeline permite elegir un pipeline y devuelve su ID.
func PickPipeline(client *azdevops.Client, title string) (int, error) {
	var defs []models.BuildDefinition
	err := ui.Spinner("Cargando pipelines...", func() (err error) {
		defs, err = pipeline.ListDefinitions(client, "", "")
		return err
	})
	if err != nil {
		return 0, err
	}
	opts := make([]ui.Option[int], 0, len(defs))
	for _, d := range defs {
		if d.ID == nil {
			continue
		}
		label := fmt.Sprintf("%s%s  #%d", FolderPrefix(d.Path), Str(d.Name), *d.ID)
		if d.LatestCompletedBuild != nil {
			label += "  · " + d.LatestCompletedBuild.Result
		}
		opts = append(opts, ui.Option[int]{Label: label, Value: int(*d.ID)})
	}
	return ui.Select(title, opts)
}

// PickBuild permite elegir una ejecución reciente (de un pipeline o de todo el proyecto).
func PickBuild(client *azdevops.Client, definitionID int, title string) (int, error) {
	var builds []models.Build
	err := ui.Spinner("Cargando ejecuciones...", func() (err error) {
		builds, err = pipeline.ListBuilds(client, definitionID, 30, "")
		return err
	})
	if err != nil {
		return 0, err
	}
	opts := make([]ui.Option[int], len(builds))
	for i, b := range builds {
		state := b.Result
		if state == "" {
			state = b.Status
		}
		opts[i] = ui.Option[int]{
			Label: fmt.Sprintf("#%d  %s  %s  %s  %s", b.ID, b.Definition.Name, b.BuildNumber, ShortBranch(b.SourceBranch), state),
			Value: b.ID,
		}
	}
	return ui.Select(title, opts)
}

// ResolveIntFlag lee un flag entero; si es 0 y hay terminal, usa pick.
func ResolveIntFlag(cmd *cobra.Command, flag string, pick func() (int, error)) (int, error) {
	v, _ := cmd.Flags().GetInt(flag)
	if v != 0 {
		return v, nil
	}
	if !ui.Interactive() {
		return 0, ui.MissingError(flag)
	}
	return pick()
}

// ResolveStringFlag lee un flag de texto; si está vacío y hay terminal, usa pick.
func ResolveStringFlag(cmd *cobra.Command, flag string, pick func() (string, error)) (string, error) {
	v, _ := cmd.Flags().GetString(flag)
	if v != "" {
		return v, nil
	}
	if !ui.Interactive() {
		return "", ui.MissingError(flag)
	}
	return pick()
}

// ResolveSliceFlag lee un flag de lista; si está vacío y hay terminal, usa pick.
func ResolveSliceFlag(cmd *cobra.Command, flag string, pick func() ([]string, error)) ([]string, error) {
	v, _ := cmd.Flags().GetStringSlice(flag)
	if len(v) > 0 {
		return v, nil
	}
	if !ui.Interactive() {
		return nil, ui.MissingError(flag)
	}
	return pick()
}

// CompleteVariableGroups autocompleta nombres de Variable Groups en la shell.
func CompleteVariableGroups(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := NewClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	groups, err := vg.ListVariableGroups(client, toComplete+"*")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// CompletePipelineIDs autocompleta IDs de pipelines mostrando su nombre.
func CompletePipelineIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := NewClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defs, err := pipeline.ListDefinitions(client, "", "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var out []string
	for _, d := range defs {
		if d.ID != nil {
			out = append(out, strconv.FormatInt(*d.ID, 10)+"\t"+Str(d.Name))
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func Str(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func FolderPrefix(path *string) string {
	p := Str(path)
	if p == "" || p == "\\" {
		return ""
	}
	return p[1:] + "\\"
}

func ShortBranch(ref string) string {
	if len(ref) > len("refs/heads/") && ref[:len("refs/heads/")] == "refs/heads/" {
		return ref[len("refs/heads/"):]
	}
	return ref
}
