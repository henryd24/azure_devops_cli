package pipeline

import (
	"fmt"
	"strconv"
	"time"

	"azuredevops/azdevops"

	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var runsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Lista las ejecuciones recientes (de un pipeline o de todo el proyecto)",
	Example: `  azdevops pipelines runs
  azdevops pipelines runs --id 123 --branch main --top 5`,
	RunE: func(c *cobra.Command, args []string) error {
		out, err := cmd.Output("table")
		if err != nil {
			return err
		}
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		id, _ := c.Flags().GetInt("id")
		if name, _ := c.Flags().GetString("name"); name != "" {
			if id, err = resolvePipelineID(c, client); err != nil {
				return err
			}
		}
		top, _ := c.Flags().GetInt("top")
		branch, _ := c.Flags().GetString("branch")

		var builds []models.Build
		if err := ui.Spinner("Cargando ejecuciones...", func() (err error) {
			builds, err = pipeline.ListBuilds(client, id, top, branch)
			return err
		}); err != nil {
			return err
		}
		if out == "json" {
			if builds == nil {
				builds = []models.Build{}
			}
			return ui.PrintJSON(builds)
		}
		if len(builds) == 0 {
			ui.Info("No hay ejecuciones.")
			return nil
		}
		printBuildsTable(builds)
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Muestra el estado y las etapas de una ejecución",
	Example: `  azdevops pipelines status --build-id 4567
  azdevops pipelines status --build-id 4567 --watch`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		buildID, err := resolveBuildID(c, client)
		if err != nil {
			return err
		}
		build, err := pipeline.GetBuildByID(client, buildID)
		if err != nil {
			return err
		}
		if watch, _ := c.Flags().GetBool("watch"); watch && build.Status != "completed" {
			ui.Link(build.Links.Web.Href)
			return waitAndReport(client, buildID, 10*time.Second, 0)
		}
		if cmd.WantsJSON() {
			return ui.PrintJSON(build)
		}
		printBuildsTable([]models.Build{*build})
		fmt.Fprintln(ui.Out)
		if tl, err := pipeline.GetTimeline(client, buildID); err == nil {
			printTimeline(tl, false)
			if build.Result == "failed" {
				printFailures(tl)
			}
		}
		ui.Link(build.Links.Web.Href)
		return nil
	},
}

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancela una ejecución en curso",
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		buildID, err := cmd.ResolveIntFlag(c, "build-id", func() (int, error) {
			builds, err := pipeline.ListBuilds(client, 0, 50, "")
			if err != nil {
				return 0, err
			}
			var opts []ui.Option[int]
			for _, b := range builds {
				if b.Status != "completed" {
					opts = append(opts, ui.Option[int]{
						Label: fmt.Sprintf("#%d  %s  %s  %s", b.ID, b.Definition.Name, cmd.ShortBranch(b.SourceBranch), b.Status),
						Value: b.ID,
					})
				}
			}
			if len(opts) == 0 {
				return 0, fmt.Errorf("no hay ejecuciones en curso")
			}
			return ui.Select("Ejecución a cancelar", opts)
		})
		if err != nil {
			return err
		}
		yes, _ := c.Flags().GetBool("yes")
		if err := cmd.Confirm(fmt.Sprintf("¿Cancelar la ejecución %d?", buildID), yes); err != nil {
			return err
		}
		if err := pipeline.CancelBuild(client, buildID); err != nil {
			return err
		}
		ui.Success("Cancelación solicitada para la ejecución %d", buildID)
		return nil
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Muestra los logs de una ejecución",
	Long: `Sin --log-id ni --failed muestra las etapas/tareas con su ID de log
(en una terminal te deja elegir cuál ver).`,
	Example: `  azdevops pipelines logs --build-id 4567 --failed
  azdevops pipelines logs --build-id 4567 --log-id 12`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		buildID, err := resolveBuildID(c, client)
		if err != nil {
			return err
		}
		if logID, _ := c.Flags().GetInt("log-id"); logID != 0 {
			return printLog(client, buildID, logID, "")
		}
		tl, err := pipeline.GetTimeline(client, buildID)
		if err != nil {
			return err
		}
		if failed, _ := c.Flags().GetBool("failed"); failed {
			found := false
			for _, r := range sortedRecords(tl) {
				if r.Type == "Task" && r.Result == "failed" && r.Log != nil {
					found = true
					if err := printLog(client, buildID, r.Log.ID, r.Name); err != nil {
						return err
					}
				}
			}
			if !found {
				ui.Info("No hay tareas fallidas en la ejecución %d", buildID)
			}
			return nil
		}

		if !ui.Interactive() {
			printTimelineWithTasks(tl)
			return nil
		}
		var opts []ui.Option[int]
		for _, r := range sortedRecords(tl) {
			if r.Log == nil {
				continue
			}
			state := r.Result
			if state == "" {
				state = r.State
			}
			opts = append(opts, ui.Option[int]{Label: fmt.Sprintf("[%s] %s · %s", r.Type, r.Name, state), Value: r.Log.ID})
		}
		logID, err := ui.Select("Log a mostrar", opts)
		if err != nil {
			return err
		}
		return printLog(client, buildID, logID, "")
	},
}

func printTimelineWithTasks(tl *models.Timeline) {
	var rows [][]string
	for _, r := range sortedRecords(tl) {
		if r.Log == nil {
			continue
		}
		state := r.Result
		if state == "" {
			state = r.State
		}
		rows = append(rows, []string{strconv.Itoa(r.Log.ID), r.Type, r.Name, ui.Status(state)})
	}
	ui.Table([]string{"log", "tipo", "nombre", "estado"}, rows)
}

func printLog(client *azdevops.Client, buildID, logID int, title string) error {
	text, err := pipeline.GetLog(client, buildID, logID)
	if err != nil {
		return err
	}
	if title != "" {
		fmt.Fprintf(ui.Out, "\n===== %s (log %d) =====\n", title, logID)
	}
	fmt.Fprintln(ui.Out, text)
	return nil
}

func init() {
	addPipelineFlags(runsCmd, "")
	runsCmd.Flags().Int("top", 20, "Número máximo de ejecuciones")
	runsCmd.Flags().StringP("branch", "b", "", "Filtrar por rama")

	for _, c := range []*cobra.Command{statusCmd, cancelCmd, logsCmd} {
		c.Flags().Int("build-id", 0, "ID de la ejecución (build)")
	}
	statusCmd.Flags().BoolP("watch", "w", false, "Seguir la ejecución hasta que termine")
	cancelCmd.Flags().BoolP("yes", "y", false, "Confirmar sin preguntar")
	logsCmd.Flags().Int("log-id", 0, "ID del log a mostrar")
	logsCmd.Flags().Bool("failed", false, "Mostrar solo los logs de las tareas fallidas")

	cmd.Pipelines.AddCommand(runsCmd, statusCmd, cancelCmd, logsCmd)
}
