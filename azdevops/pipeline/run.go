package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type RunOptions struct {
	Branch string
	Params map[string]string
	Vars   map[string]models.BuildVariable
}

// RunPipeline inicia una ejecución del pipeline.
func RunPipeline(client *azdevops.Client, pipelineID int, opts RunOptions) (*models.PipelineRun, error) {
	payload := models.BuildRunPayload{}
	if len(opts.Params) > 0 {
		payload.TemplateParameters = opts.Params
	}
	if len(opts.Vars) > 0 {
		payload.Variables = opts.Vars
	}
	if opts.Branch != "" {
		ref := opts.Branch
		if !strings.HasPrefix(ref, "refs/") {
			ref = "refs/heads/" + ref
		}
		payload.Resources = &models.RunResources{
			Repositories: map[string]models.RunRepository{"self": {RefName: ref}},
		}
	}

	var run models.PipelineRun
	u := client.ProjectURL(fmt.Sprintf("pipelines/%d/runs", pipelineID), url.Values{"api-version": {"7.1"}})
	if err := client.Do("POST", u, payload, &run); err != nil {
		return nil, fmt.Errorf("error al iniciar el pipeline %d: %w", pipelineID, err)
	}
	return &run, nil
}

// WaitForBuild consulta el build cada interval hasta que termina o se agota timeout
// (0 = sin límite). onUpdate (opcional) se invoca en cada consulta.
func WaitForBuild(client *azdevops.Client, buildID int, interval, timeout time.Duration, onUpdate func(*models.Build)) (*models.Build, error) {
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	failures := 0
	for {
		build, err := GetBuildByID(client, buildID)
		if err != nil {
			// Toleramos errores transitorios aislados mientras esperamos.
			failures++
			if failures >= 5 {
				return nil, fmt.Errorf("no se pudo obtener el estado del build %d: %w", buildID, err)
			}
		} else {
			failures = 0
			if onUpdate != nil {
				onUpdate(build)
			}
			if build.Status == "completed" {
				return build, nil
			}
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return build, fmt.Errorf("se agotó el tiempo de espera (%s) para el build %d", timeout, buildID)
		}
		time.Sleep(interval)
	}
}
