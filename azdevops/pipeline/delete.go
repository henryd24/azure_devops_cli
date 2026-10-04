package pipeline

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// DeletePipeline elimina un pipeline junto con sus retenciones. Devuelve cuántas
// retenciones se eliminaron.
func DeletePipeline(client *azdevops.Client, pipelineID int) (int, error) {
	leases, err := getRetentionLeases(client, pipelineID)
	if err != nil {
		return 0, fmt.Errorf("no se pudieron obtener las retenciones: %w", err)
	}

	if len(leases) > 0 {
		ids := make([]string, len(leases))
		for i, lease := range leases {
			ids[i] = strconv.Itoa(lease.LeaseID)
		}
		q := url.Values{"ids": {strings.Join(ids, ",")}, "api-version": {"7.1-preview.1"}}
		if err := client.Do("DELETE", client.ProjectURL("build/retention/leases", q), nil, nil); err != nil {
			return 0, fmt.Errorf("error al eliminar las retenciones: %w", err)
		}
	}

	u := client.ProjectURL(fmt.Sprintf("build/definitions/%d", pipelineID), url.Values{"api-version": {"7.1"}})
	if err := client.Do("DELETE", u, nil, nil); err != nil {
		return len(leases), fmt.Errorf("error al eliminar el pipeline %d: %w", pipelineID, err)
	}
	return len(leases), nil
}

func getRetentionLeases(client *azdevops.Client, pipelineID int) ([]models.RetentionLease, error) {
	q := url.Values{"definitionId": {strconv.Itoa(pipelineID)}, "api-version": {"7.1-preview.1"}}
	var result struct {
		Value []models.RetentionLease `json:"value"`
	}
	if err := client.Do("GET", client.ProjectURL("build/retention/leases", q), nil, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}
