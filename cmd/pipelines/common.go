package pipeline

import (
	"fmt"
	"strconv"

	"azuredevops/azdevops"
	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

// addPipelineFlags registra --id y --name para identificar un pipeline.
func addPipelineFlags(c *cobra.Command, verb string) {
	c.Flags().IntP("id", "i", 0, "ID del pipeline "+verb)
	c.Flags().StringP("name", "n", "", "Nombre exacto del pipeline "+verb+" (alternativa a --id)")
	_ = c.RegisterFlagCompletionFunc("id", cmd.CompletePipelineIDs)
}

// resolvePipelineID usa --id, luego --name y por último un selector interactivo.
func resolvePipelineID(c *cobra.Command, client *azdevops.Client) (int, error) {
	if id, _ := c.Flags().GetInt("id"); id != 0 {
		return id, nil
	}
	if name, _ := c.Flags().GetString("name"); name != "" {
		defs, err := pipeline.GetBuildDefinitionByName(client, name)
		if err != nil {
			return 0, err
		}
		var matches []models.BuildDefinition
		for _, d := range defs {
			if d.Name != nil && *d.Name == name && d.ID != nil {
				matches = append(matches, d)
			}
		}
		switch len(matches) {
		case 0:
			return 0, fmt.Errorf("no se encontró el pipeline '%s'", name)
		case 1:
			return int(*matches[0].ID), nil
		default:
			return 0, fmt.Errorf("hay %d pipelines llamados '%s' (en distintas carpetas); usa --id", len(matches), name)
		}
	}
	if !ui.Interactive() {
		return 0, fmt.Errorf("debes indicar el pipeline con --id o --name")
	}
	return cmd.PickPipeline(client, "Pipeline")
}

// resolveBuildID usa --build-id o un selector de ejecuciones recientes.
func resolveBuildID(c *cobra.Command, client *azdevops.Client) (int, error) {
	return cmd.ResolveIntFlag(c, "build-id", func() (int, error) {
		return cmd.PickBuild(client, 0, "Ejecución")
	})
}

func buildState(b *models.Build) string {
	if b.Status == "completed" {
		return ui.Status(b.Result)
	}
	return ui.Status(b.Status)
}

func printBuildsTable(builds []models.Build) {
	rows := make([][]string, len(builds))
	for i, b := range builds {
		rows[i] = []string{
			strconv.Itoa(b.ID), b.Definition.Name, b.BuildNumber, cmd.ShortBranch(b.SourceBranch),
			buildState(&b), b.RequestedFor.DisplayName, ui.FormatTime(b.QueueTime), ui.FormatDuration(b.StartTime, b.FinishTime),
		}
	}
	ui.Table([]string{"id", "pipeline", "número", "rama", "estado", "solicitado por", "encolado", "duración"}, rows)
}

// printTimeline muestra etapas, jobs y tareas fallidas de una ejecución.
func printTimeline(timeline *models.Timeline, onlyFailed bool) {
	var rows [][]string
	for _, r := range sortedRecords(timeline) {
		if r.Type != "Stage" && r.Type != "Job" && r.Type != "Task" {
			continue
		}
		if onlyFailed && r.Result != "failed" {
			continue
		}
		if r.Type == "Task" && !onlyFailed && r.Result != "failed" {
			continue
		}
		state := r.Result
		if state == "" {
			state = r.State
		}
		logID := "-"
		if r.Log != nil {
			logID = strconv.Itoa(r.Log.ID)
		}
		indent := map[string]string{"Stage": "", "Job": "  ", "Task": "    "}[r.Type]
		rows = append(rows, []string{indent + r.Name, r.Type, ui.Status(state), logID})
	}
	if len(rows) == 0 {
		return
	}
	ui.Table([]string{"nombre", "tipo", "estado", "log"}, rows)
}

func sortedRecords(t *models.Timeline) []models.TimelineRecord {
	// Ordena jerárquicamente: etapa -> job -> tarea, respetando "order".
	children := map[string][]models.TimelineRecord{}
	for _, r := range t.Records {
		children[r.ParentID] = append(children[r.ParentID], r)
	}
	var out []models.TimelineRecord
	var walk func(parent string)
	walk = func(parent string) {
		list := children[parent]
		for i := 1; i < len(list); i++ {
			for j := i; j > 0 && list[j].Order < list[j-1].Order; j-- {
				list[j], list[j-1] = list[j-1], list[j]
			}
		}
		for _, r := range list {
			out = append(out, r)
			walk(r.ID)
		}
	}
	walk("")
	return out
}

// printFailures muestra los errores reportados por las tareas fallidas.
func printFailures(timeline *models.Timeline) {
	for _, r := range sortedRecords(timeline) {
		if r.Type != "Task" || r.Result != "failed" {
			continue
		}
		logInfo := ""
		if r.Log != nil {
			logInfo = fmt.Sprintf(" (log %d)", r.Log.ID)
		}
		ui.Errorf("Tarea fallida: %s%s", ui.Bold(r.Name), logInfo)
		for _, issue := range r.Issues {
			if issue.Type == "error" {
				fmt.Fprintf(ui.Err, "    %s\n", issue.Message)
			}
		}
	}
}

// currentActivity resume qué etapas/jobs están en ejecución.
func currentActivity(timeline *models.Timeline) string {
	var stage, job string
	for _, r := range sortedRecords(timeline) {
		if r.State != "inProgress" {
			continue
		}
		switch r.Type {
		case "Stage":
			stage = r.Name
		case "Job":
			job = r.Name
		}
	}
	switch {
	case stage != "" && job != "":
		return stage + " › " + job
	case stage != "":
		return stage
	}
	return job
}
