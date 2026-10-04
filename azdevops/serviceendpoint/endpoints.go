package serviceendpoint

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"net/url"
	"sort"
)

// ListEndpoints lista las conexiones de servicio del proyecto. endpointType (p. ej. "github")
// es opcional.
func ListEndpoints(client *azdevops.Client, endpointType string) ([]models.ServiceEndpoint, error) {
	q := url.Values{"api-version": {"7.1"}}
	if endpointType != "" {
		q.Set("type", endpointType)
	}
	var result struct {
		Value []models.ServiceEndpoint `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("serviceendpoint/endpoints", q), nil, &result); err != nil {
		return nil, err
	}
	sort.Slice(result.Value, func(i, j int) bool { return result.Value[i].Name < result.Value[j].Name })
	return result.Value, nil
}
