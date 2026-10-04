package agent

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// ListPools lista los agent pools de la organización.
func ListPools(client *azdevops.Client) ([]models.AgentPool, error) {
	var result struct {
		Value []models.AgentPool `json:"value"`
	}
	if err := client.Do("GET", client.OrgURL("distributedtask/pools", url.Values{"api-version": {"7.1"}}), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar los agent pools: %w", err)
	}
	sort.Slice(result.Value, func(i, j int) bool { return result.Value[i].Name < result.Value[j].Name })
	return result.Value, nil
}

// FindPool busca un pool por nombre (sin distinguir mayúsculas) o ID.
func FindPool(client *azdevops.Client, nameOrID string) (*models.AgentPool, error) {
	pools, err := ListPools(client)
	if err != nil {
		return nil, err
	}
	for i, p := range pools {
		if strings.EqualFold(p.Name, nameOrID) || fmt.Sprint(p.ID) == nameOrID {
			return &pools[i], nil
		}
	}
	return nil, fmt.Errorf("no se encontró el agent pool '%s'", nameOrID)
}

// ListAgents lista los agentes de un pool con su trabajo actual y el último completado.
func ListAgents(client *azdevops.Client, poolID int) ([]models.Agent, error) {
	q := url.Values{"api-version": {"7.1"}, "includeAssignedRequest": {"true"}, "includeLastCompletedRequest": {"true"}}
	var result struct {
		Value []models.Agent `json:"value"`
	}
	if err := client.Do("GET", client.OrgURL(fmt.Sprintf("distributedtask/pools/%d/agents", poolID), q), nil, &result); err != nil {
		return nil, fmt.Errorf("error al listar los agentes: %w", err)
	}
	sort.Slice(result.Value, func(i, j int) bool { return result.Value[i].Name < result.Value[j].Name })
	return result.Value, nil
}

// SetAgentEnabled habilita o deshabilita un agente.
func SetAgentEnabled(client *azdevops.Client, poolID, agentID int, enabled bool) (*models.Agent, error) {
	var updated models.Agent
	body := map[string]any{"id": agentID, "enabled": enabled}
	u := client.OrgURL(fmt.Sprintf("distributedtask/pools/%d/agents/%d", poolID, agentID), url.Values{"api-version": {"7.1"}})
	if err := client.Do("PATCH", u, body, &updated); err != nil {
		return nil, fmt.Errorf("error al actualizar el agente: %w", err)
	}
	return &updated, nil
}
