package pipeline

import (
	"fmt"
	"os"
	"strings"
	"time"

	"azuredevops/azdevops"
	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"
	"azuredevops/internal/ui"
	"azuredevops/models"

	"github.com/spf13/cobra"
)

var runPipelineCmd = &cobra.Command{
	Use:   "run",
	Short: "Ejecuta un pipeline y opcionalmente espera a que finalice",
	Long: `Ejecuta un pipeline. Con --wait muestra el progreso en vivo y, al terminar,
devuelve un código de salida distinto de 0 si la ejecución falló o fue cancelada
(útil en scripts y CI). Si falla, se listan las tareas con error.`,
	Example: `  azdevops pipelines run --id 123 --wait
  azdevops pipelines run --name MiPipeline --branch feature/x --wait --timeout 30m
  azdevops pipelines run --id 123 --param imageTag=1.2.3 --var deployEnv=staging --var secret:apiKey=xxx
  azdevops pipelines run   # modo guiado`,
	RunE: func(c *cobra.Command, args []string) error {
		client, err := cmd.NewClient()
		if err != nil {
			return err
		}
		f := c.Flags()
		guided := !f.Changed("id") && !f.Changed("name") && ui.Interactive()

		id, err := resolvePipelineID(c, client)
		if err != nil {
			return err
		}
		wait, _ := f.GetBool("wait")
		branch, _ := f.GetString("branch")
		paramEntries, _ := f.GetStringSlice("param")
		varEntries, _ := f.GetStringSlice("var")
		timeout, _ := f.GetDuration("timeout")
		interval, _ := f.GetDuration("interval")

		if guided {
			if branch, err = ui.Input("Rama", "Vacío para usar la rama por defecto del pipeline", branch, nil); err != nil {
				return err
			}
			p, err := ui.Input("Parámetros (opcional)", "Formato clave=valor separados por coma", "", nil)
			if err != nil {
				return err
			}
			for _, entry := range strings.Split(p, ",") {
				if entry = strings.TrimSpace(entry); entry != "" {
					paramEntries = append(paramEntries, entry)
				}
			}
			if !f.Changed("wait") {
				if wait, err = ui.Confirm("¿Esperar a que termine y ver el progreso?", true); err != nil {
					return err
				}
			}
		}

		params, err := cmd.ParseParams(paramEntries)
		if err != nil {
			return err
		}
		vars, err := cmd.ParseBuildVariables(varEntries)
		if err != nil {
			return err
		}

		run, err := pipeline.RunPipeline(client, id, pipeline.RunOptions{Branch: branch, Params: params, Vars: vars})
		if err != nil {
			return err
		}
		ui.Success("Pipeline iniciado (ejecución %d: %s)", run.ID, run.Name)
		ui.Link(run.Links.Web.Href)

		if !wait {
			if cmd.WantsData() {
				return cmd.Print(run)
			}
			return nil
		}
		return waitAndReport(client, run.ID, interval, timeout)
	},
}

// waitAndReport espera a que termine el build mostrando el progreso y devuelve error
// si terminó en fallo o cancelación.
func waitAndReport(client *azdevops.Client, buildID int, interval, timeout time.Duration) error {
	progress := ui.StartProgress("Esperando a que el pipeline inicie...")
	start := time.Now()
	plain := !ui.IsTerminal(os.Stderr)
	last := ""
	build, err := pipeline.WaitForBuild(client, buildID, interval, timeout, func(b *models.Build) {
		activity := ""
		if b.Status == "inProgress" {
			if tl, err := pipeline.GetTimeline(client, buildID); err == nil {
				activity = currentActivity(tl)
			}
		}
		title := fmt.Sprintf("%s · %s", b.Status, time.Since(start).Round(time.Second))
		if activity != "" {
			title += " · " + activity
		}
		progress.SetTitle(title)
		// Sin terminal (p. ej. en CI) se imprime una línea por cada cambio de estado.
		if plain && b.Status+activity != last {
			last = b.Status + activity
			ui.Info("%s", title)
		}
	})
	progress.Stop()
	if err != nil {
		return err
	}

	if cmd.WantsData() {
		_ = cmd.Print(build)
	}
	duration := ui.FormatDuration(build.StartTime, build.FinishTime)
	switch build.Result {
	case "succeeded":
		ui.Success("Ejecución %d finalizada: %s (%s)", build.ID, ui.Status(build.Result), duration)
		return nil
	case "partiallySucceeded":
		ui.Warn("Ejecución %d finalizada: %s (%s)", build.ID, ui.Status(build.Result), duration)
		return nil
	}

	ui.Errorf("Ejecución %d finalizada: %s (%s)", build.ID, ui.Status(build.Result), duration)
	if tl, err := pipeline.GetTimeline(client, buildID); err == nil {
		printFailures(tl)
		ui.Info("Ver logs: azdevops pipelines logs --build-id %d --failed", build.ID)
	}
	return fmt.Errorf("el pipeline terminó con resultado '%s'", build.Result)
}

func init() {
	addPipelineFlags(runPipelineCmd, "a ejecutar")
	f := runPipelineCmd.Flags()
	f.BoolP("wait", "w", false, "Esperar a que el pipeline finalice mostrando el progreso")
	f.StringP("branch", "b", "", "Rama a ejecutar (por defecto la del pipeline)")
	f.StringSliceP("param", "p", nil, "Parámetros del template YAML: clave=valor (se puede repetir)")
	f.StringSliceP("var", "v", nil, "Variables de ejecución: clave=valor o secret:clave=valor")
	f.Duration("timeout", 0, "Tiempo máximo de espera con --wait (p. ej. 30m). 0 = sin límite")
	f.Duration("interval", 10*time.Second, "Intervalo entre consultas de estado con --wait")
	cmd.Pipelines.AddCommand(runPipelineCmd)
}
