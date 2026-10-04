package organization

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
)

// GetProject devuelve la información del proyecto configurado en el cliente.
func GetProject(client *azdevops.Client) (*models.Project, error) {
	var result models.Project
	u := client.OrgURL("projects/"+url.PathEscape(client.Project), url.Values{"api-version": {"7.1"}})
	if err := client.Do("GET", u, nil, &result); err != nil {
		return nil, fmt.Errorf("no se pudo obtener el proyecto '%s': %w", client.Project, err)
	}
	return &result, nil
}

// ListProjects devuelve los proyectos de la organización.
func ListProjects(client *azdevops.Client) ([]models.Project, error) {
	var result struct {
		Value []models.Project `json:"value"`
	}
	u := client.OrgURL("projects", url.Values{"api-version": {"7.1"}, "$top": {"500"}})
	if err := client.Do("GET", u, nil, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetProjectDescriptor devuelve el descriptor de Graph del proyecto.
func GetProjectDescriptor(client *azdevops.Client) (string, error) {
	project, err := GetProject(client)
	if err != nil {
		return "", err
	}
	var result models.GraphDescriptor
	if err := client.Do("GET", client.VSSPSURL("graph/descriptors/"+project.ID, nil), nil, &result); err != nil {
		return "", fmt.Errorf("error al obtener el descriptor del proyecto: %w", err)
	}
	return result.ID, nil
}
