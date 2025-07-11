package variable_group

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func CreateVariableGroup(client *azdevops.Client, name string, variables map[string]models.VariableVal, description string) (bool, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups?projectIds={projectIds}&api-version=7.1-preview.2", client.Org, client.Project)
	group := models.VariableGroupById{
		Name:                           name,
		Type:                           "Vsts",
		Project:                        client.Project,
		Variables:                      variables,
		VariableGroupProjectReferences: models.ConstructVariableGroupProjectReferences(client.Project, name, description),
	}

	payload, err := json.Marshal(group)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", client.AuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("CreateVariableGroup error: %s", string(body))
	}

	var respJSON models.VariableGroup
	if err := json.NewDecoder(resp.Body).Decode(&respJSON); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("✔ Grupo de variables creado (ID: %d)\n", respJSON.Id)
	fmt.Printf("🔗 https://dev.azure.com/%s/%s/_library?itemType=variableGroups&view=VariableGroupView&variableGroupId=%d\n", client.Org, client.Project, respJSON.Id)
	return true, nil
}
