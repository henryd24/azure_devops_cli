package organization

import (
	"azuredevops/azdevops"
	"azuredevops/models"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetOrganizationInfo(client *azdevops.Client) (*models.Project, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/_apis/projects/%s?api-version=7.1", client.Org, client.Project)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.Project
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func GetProjectDescriptor(client *azdevops.Client) (string, error) {
	project, err := GetOrganizationInfo(client)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://vssps.dev.azure.com/%s/_apis/graph/descriptors/%s", client.Org, project.ID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", client.AuthHeader())

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result models.GraphDescriptor
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error al decodificar la respuesta del descriptor: %w", err)
	}

	return result.ID, nil
}
