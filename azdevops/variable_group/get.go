package variable_group

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetVariableGroupByName(client *azdevops.Client, name string) ([]models.VariableGroup, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups?api-version=7.1-preview.2&groupName=%s", client.Org, client.Project, name)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
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

func GetVariableGroupById(client *azdevops.Client, id int) (*models.VariableGroup, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/distributedtask/variablegroups/%d?api-version=7.1-preview.2", client.Org, client.Project, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.VariableGroup
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
