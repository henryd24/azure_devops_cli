package serviceendpoint

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"sort"
	"strconv"
)

const apiVersion = "7.1"

// ListEndpoints lista las conexiones de servicio del proyecto. endpointType (p. ej. "github")
// es opcional.
func ListEndpoints(client *azdevops.Client, endpointType string) ([]models.ServiceEndpoint, error) {
	q := url.Values{"api-version": {apiVersion}}
	if endpointType != "" {
		q.Set("type", endpointType)
	}
	var result struct {
		Value []models.ServiceEndpoint `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("serviceendpoint/endpoints", q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar las conexiones de servicio: %w", err)
	}
	sort.Slice(result.Value, func(i, j int) bool { return result.Value[i].Name < result.Value[j].Name })
	return result.Value, nil
}

// GetEndpoint obtiene una conexión de servicio por su ID.
func GetEndpoint(client *azdevops.Client, id string) (*models.ServiceEndpoint, error) {
	var ep models.ServiceEndpoint
	u := client.ProjectURL("serviceendpoint/endpoints/"+url.PathEscape(id), url.Values{"api-version": {apiVersion}})
	if err := client.Do("GET", u, nil, &ep); err != nil {
		return nil, err
	}
	if ep.ID == "" {
		return nil, fmt.Errorf("no se encontró la conexión de servicio '%s'", id)
	}
	return &ep, nil
}

// GetEndpointByName busca una conexión de servicio por su nombre exacto.
func GetEndpointByName(client *azdevops.Client, name string) (*models.ServiceEndpoint, error) {
	var result struct {
		Value []models.ServiceEndpoint `json:"value"`
	}
	q := url.Values{"api-version": {apiVersion}, "endpointNames": {name}}
	if err := client.Do("GET", client.ProjectURL("serviceendpoint/endpoints", q), nil, &result); err != nil {
		return nil, err
	}
	for i := range result.Value {
		if result.Value[i].Name == name {
			return &result.Value[i], nil
		}
	}
	return nil, fmt.Errorf("no se encontró la conexión de servicio '%s'", name)
}

// ExecutionHistory devuelve las ejecuciones de pipelines que usaron la conexión.
func ExecutionHistory(client *azdevops.Client, id string, top int) ([]models.ServiceEndpointExecution, error) {
	var result struct {
		Value []struct {
			Data models.ServiceEndpointExecution `json:"data"`
		} `json:"value"`
	}
	q := url.Values{"api-version": {apiVersion}, "top": {strconv.Itoa(top)}}
	u := client.ProjectURL("serviceendpoint/"+url.PathEscape(id)+"/executionhistory", q)
	if err := client.Do("GET", u, nil, &result); err != nil {
		return nil, fmt.Errorf("error al obtener el historial: %w", err)
	}
	out := make([]models.ServiceEndpointExecution, len(result.Value))
	for i, v := range result.Value {
		out[i] = v.Data
	}
	return out, nil
}

// ShareEndpoint comparte una conexión de servicio con otros proyectos de la organización.
func ShareEndpoint(client *azdevops.Client, id string, refs []models.ServiceEndpointProjectReference) error {
	u := client.OrgURL("serviceendpoint/endpoints/"+url.PathEscape(id), url.Values{"api-version": {apiVersion}})
	if err := client.Do("PATCH", u, refs, nil); err != nil {
		return fmt.Errorf("error al compartir la conexión: %w", err)
	}
	return nil
}
