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

func (c *Client) CreateVariableGroup(group models.VariableGroup) error {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups?api-version=7.1-preview.2", c.Org, c.Project)
	payload, _ := json.Marshal(group)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", c.authHeader())
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("CreateVariableGroup error: %s", string(body))
	}
	return nil
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

func (c *Client) AddVariablesToGroup(variableGroup models.VariableGroup, variables map[string]models.VariableVal) (bool, error) {
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
		VariableGroupProjectReferences: models.ConstructVariableGroupProjectReferences(c.Project, variableGroup.Name),
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

	fmt.Printf("Variables agregadas satisfactoriamente al grupo: %d\n", variableGroup.Id)
	fmt.Printf("Verifica en la url: https://dev.azure.com/%s/%s/_library?itemType=variableGroups&view=VariableGroupView&variableGroupId=%d ", c.Org, c.Project, variableGroup.Id)
	return true, nil
}
