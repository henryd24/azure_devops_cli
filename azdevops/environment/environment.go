package environment

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"sort"
	"strconv"
)

// ListEnvironments lista los environments del proyecto.
func ListEnvironments(client *azdevops.Client) ([]models.Environment, error) {
	var result struct {
		Value []models.Environment `json:"value"`
	}
	q := url.Values{"api-version": {"7.1"}, "$top": {"500"}}
	if err := client.Do("GET", client.ProjectURL("pipelines/environments", q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar los environments: %w", err)
	}
	sort.Slice(result.Value, func(i, j int) bool { return result.Value[i].Name < result.Value[j].Name })
	return result.Value, nil
}

// ListDeployments devuelve los despliegues más recientes de un environment.
func ListDeployments(client *azdevops.Client, environmentID, top int) ([]models.EnvironmentDeployment, error) {
	var result struct {
		Value []models.EnvironmentDeployment `json:"value"`
	}
	q := url.Values{"api-version": {"7.1"}, "top": {strconv.Itoa(top)}}
	path := fmt.Sprintf("pipelines/environments/%d/environmentdeploymentrecords", environmentID)
	if err := client.Do("GET", client.ProjectURL(path, q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar los despliegues: %w", err)
	}
	return result.Value, nil
}

// ListApprovals lista las aprobaciones del proyecto en el estado indicado (p. ej. "pending").
func ListApprovals(client *azdevops.Client, state string) ([]models.Approval, error) {
	q := url.Values{"api-version": {"7.1"}, "$expand": {"steps"}}
	if state != "" {
		q.Set("state", state)
	}
	var result struct {
		Value []models.Approval `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("pipelines/approvals", q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar las aprobaciones: %w", err)
	}
	return result.Value, nil
}

// UpdateApprovals aprueba o rechaza (status "approved" / "rejected") las aprobaciones indicadas.
func UpdateApprovals(client *azdevops.Client, ids []string, status, comment string) ([]models.Approval, error) {
	type update struct {
		ApprovalID string `json:"approvalId"`
		Status     string `json:"status"`
		Comment    string `json:"comment,omitempty"`
	}
	body := make([]update, len(ids))
	for i, id := range ids {
		body[i] = update{ApprovalID: id, Status: status, Comment: comment}
	}
	var result struct {
		Value []models.Approval `json:"value"`
	}
	if err := client.Do("PATCH", client.ProjectURL("pipelines/approvals", url.Values{"api-version": {"7.1"}}), body, &result); err != nil {
		return nil, fmt.Errorf("error al actualizar las aprobaciones: %w", err)
	}
	return result.Value, nil
}
