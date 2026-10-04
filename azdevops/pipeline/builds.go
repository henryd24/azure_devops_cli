package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// GetBuildByID obtiene una ejecución (build) por su ID.
func GetBuildByID(client *azdevops.Client, buildID int) (*models.Build, error) {
	var build models.Build
	u := client.ProjectURL(fmt.Sprintf("build/builds/%d", buildID), url.Values{"api-version": {"7.1"}})
	if err := client.Do("GET", u, nil, &build); err != nil {
		return nil, err
	}
	return &build, nil
}

// ListBuilds lista las ejecuciones más recientes. definitionID y branch son opcionales.
func ListBuilds(client *azdevops.Client, definitionID, top int, branch string) ([]models.Build, error) {
	q := url.Values{"api-version": {"7.1"}, "queryOrder": {"queueTimeDescending"}}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	if definitionID > 0 {
		q.Set("definitions", strconv.Itoa(definitionID))
	}
	if branch != "" {
		if !strings.HasPrefix(branch, "refs/") {
			branch = "refs/heads/" + branch
		}
		q.Set("branchName", branch)
	}
	var result struct {
		Value []models.Build `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("build/builds", q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar las ejecuciones: %w", err)
	}
	return result.Value, nil
}

// CancelBuild solicita la cancelación de una ejecución en curso.
func CancelBuild(client *azdevops.Client, buildID int) error {
	u := client.ProjectURL(fmt.Sprintf("build/builds/%d", buildID), url.Values{"api-version": {"7.1"}})
	if err := client.Do("PATCH", u, map[string]string{"status": "cancelling"}, nil); err != nil {
		return fmt.Errorf("error al cancelar el build %d: %w", buildID, err)
	}
	return nil
}

// GetTimeline devuelve las etapas, jobs y tareas de una ejecución.
func GetTimeline(client *azdevops.Client, buildID int) (*models.Timeline, error) {
	var timeline models.Timeline
	u := client.ProjectURL(fmt.Sprintf("build/builds/%d/timeline", buildID), url.Values{"api-version": {"7.1"}})
	if err := client.Do("GET", u, nil, &timeline); err != nil {
		return nil, fmt.Errorf("error al obtener el timeline del build %d: %w", buildID, err)
	}
	return &timeline, nil
}

// GetLog devuelve el texto de un log de una ejecución.
func GetLog(client *azdevops.Client, buildID, logID int) (string, error) {
	// Con Accept: application/json la API devuelve las líneas en un arreglo.
	var result struct {
		Value []string `json:"value"`
	}
	u := client.ProjectURL(fmt.Sprintf("build/builds/%d/logs/%d", buildID, logID), url.Values{"api-version": {"7.1"}})
	if err := client.Do("GET", u, nil, &result); err != nil {
		return "", fmt.Errorf("error al obtener el log %d del build %d: %w", logID, buildID, err)
	}
	return strings.Join(result.Value, "\n"), nil
}
