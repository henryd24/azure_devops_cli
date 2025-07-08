package azdevops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"azuredevops/models"
	"maps"
)

func (c *Client) ListVariableGroups() ([]models.VariableGroup, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups?api-version=7.1-preview.2", c.Org, c.Project)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", c.authHeader())

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.VariableGroup `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

func (c *Client) CreateVariableGroup(name string, variables map[string]models.VariableVal, description string) (bool, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups?api-version=7.1-preview.2", c.Org, c.Project)
	group := models.VariableGroupById{
		Name:                           name,
		Type:                           "Vsts",
		Project:                        c.Project,
		Variables:                      variables,
		VariableGroupProjectReferences: models.ConstructVariableGroupProjectReferences(c.Project, name, description),
	}

	payload, err := json.Marshal(group)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
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
	fmt.Printf("🔗 https://dev.azure.com/%s/%s/_library?itemType=variableGroups&view=VariableGroupView&variableGroupId=%d\n", c.Org, c.Project, respJSON.Id)
	return true, nil
}

func (c *Client) GetVariableGroupByName(name string) ([]models.VariableGroup, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups?api-version=7.1-preview.2&groupName=%s", c.Org, c.Project, name)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", c.authHeader())

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Value []models.VariableGroup `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

func (c *Client) AddVariablesToGroup(variableGroup models.VariableGroup, variables map[string]models.VariableVal, description string) (bool, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups/%d?api-version=7.1-preview.2", c.Org, c.Project, variableGroup.Id)
	mergedVars := make(map[string]models.VariableVal)
	maps.Copy(mergedVars, variableGroup.Variables)
	maps.Copy(mergedVars, variables)

	group := models.VariableGroupById{
		ID:                             variableGroup.Id,
		Name:                           variableGroup.Name,
		Type:                           "Vsts",
		Project:                        c.Project,
		Variables:                      mergedVars,
		VariableGroupProjectReferences: models.ConstructVariableGroupProjectReferences(c.Project, variableGroup.Name, description),
	}
	payload, _ := json.Marshal(group)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", c.authHeader())
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("AddVariablesToGroup error: %s", string(body))
	}

	fmt.Printf("✔ Variables agregadas satisfactoriamente al grupo: %d\n", variableGroup.Id)
	fmt.Printf("🔗 https://dev.azure.com/%s/%s/_library?itemType=variableGroups&view=VariableGroupView&variableGroupId=%d\n", c.Org, c.Project, variableGroup.Id)
	return true, nil
}
