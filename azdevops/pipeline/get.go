package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"sort"
)

// ListDefinitions lista los pipelines del proyecto. name admite comodines y folder
// filtra por carpeta (p. ej. "\\infra"). Ambos son opcionales.
func ListDefinitions(client *azdevops.Client, name, folder string) ([]models.BuildDefinition, error) {
	q := url.Values{"api-version": {"7.1"}, "includeLatestBuilds": {"true"}}
	if name != "" {
		q.Set("name", name)
	}
	if folder != "" {
		q.Set("path", folder)
	}
	var result struct {
		Value []models.BuildDefinition `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("build/definitions", q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar los pipelines: %w", err)
	}
	sort.Slice(result.Value, func(i, j int) bool {
		return deref(result.Value[i].Path)+deref(result.Value[i].Name) < deref(result.Value[j].Path)+deref(result.Value[j].Name)
	})
	return result.Value, nil
}

// GetBuildDefinitionByName busca pipelines por nombre (admite comodines).
func GetBuildDefinitionByName(client *azdevops.Client, name string) ([]models.BuildDefinition, error) {
	if name == "" {
		return nil, fmt.Errorf("el nombre del pipeline no puede estar vacío")
	}
	return ListDefinitions(client, name, "")
}

// GetBuildDefinitionByID obtiene la definición completa de un pipeline.
func GetBuildDefinitionByID(client *azdevops.Client, id int) (*models.BuildDefinition, error) {
	var definition models.BuildDefinition
	u := client.ProjectURL(fmt.Sprintf("build/definitions/%d", id), url.Values{"api-version": {"7.1"}})
	if err := client.Do("GET", u, nil, &definition); err != nil {
		return nil, fmt.Errorf("error al obtener el pipeline %d: %w", id, err)
	}
	return &definition, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
